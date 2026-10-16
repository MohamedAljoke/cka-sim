package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"slices"
	"strings"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/assistant"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/auth"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/desk"
	labagent "github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

var (
	errScoped     = domainerr.Forbidden("assistant_not_allowed", "an assistant token can't do this")
	errNoTask     = domainerr.NotFound("no_task", "the learner has no task open")
	errBadInspect = domainerr.Unprocessable("inspect_refused",
		"only read-only kubectl: "+strings.Join(inspectVerbs, ", ")+", with only these flags: "+strings.Join(inspectFlags, " ")+
			"; -o can't read a template file")
	errChatOff = domainerr.Unavailable("assistant_off", "no chat assistant is set up on this server")
)

const (
	inspectLimit   = 8 << 10
	inspectTimeout = 20 * time.Second
)

var (
	inspectVerbs = []string{"get", "describe", "logs", "events", "top", "explain", "api-resources", "version"}
	inspectFlags = []string{
		"-n", "--namespace", "-A", "--all-namespaces", "-l", "--selector", "--field-selector",
		"-o", "--output", "--show-labels", "-L", "--label-columns", "--sort-by", "--no-headers", "--show-kind",
		"--ignore-not-found", "--show-managed-fields",
		"-c", "--container", "--all-containers", "--tail", "--since", "--since-time", "-p", "--previous", "--prefix",
		"--timestamps", "--limit-bytes",
		"--recursive", "--api-version", "--namespaced", "--verbs", "--api-group", "--for", "--types",
		"--containers", "--sum", "--client",
	}
)

func sessionOnly(h httpx.Handler) httpx.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if claims, ok := auth.FromContext(r.Context()); ok && claims.Scope != "" {
			return errScoped
		}
		return h(w, r)
	}
}

func (a *api) assistantToo(h httpx.Handler) httpx.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		claims, ok := auth.FromContext(r.Context())
		if ok && claims.Scope != "" {
			if claims.Scope != auth.ScopeAssistant {
				return errScoped
			}
			d, err := a.desk(r)
			if err != nil {
				return err
			}
			if err := notDuringExam(d); err != nil {
				return err
			}
		}
		return h(w, r)
	}
}

type assistantRequest struct {
	Provider string `json:"provider"`
}

func (a *api) startAssistant(w http.ResponseWriter, r *http.Request) error {
	var req assistantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.Malformed(`the body must be JSON like {"provider": "claude"}`)
	}
	if _, _, err := a.prepareAssistant(r, req.Provider); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// Prepares again even after the POST: a socket opened on its own must not skip the settings.
func (a *api) serveAssistant(w http.ResponseWriter, r *http.Request) error {
	d, p, err := a.prepareAssistant(r, r.URL.Query().Get("provider"))
	if err != nil {
		return err
	}
	labagent.Serve(w, r, func(ctx context.Context, size labagent.Size) (labagent.Shell, error) {
		return d.OpenAssistant(ctx, p.Command(), size)
	})
	return nil
}

func (a *api) prepareAssistant(r *http.Request, id string) (*desk.Desk, assistant.Provider, error) {
	p, err := assistant.Find(a.Assistants, id)
	if err != nil {
		return nil, nil, err
	}
	d, err := a.desk(r)
	if err != nil {
		return nil, nil, err
	}
	if err := notDuringExam(d); err != nil {
		return nil, nil, err
	}
	token, err := a.Tokens.Issue(r.Context(), r)
	if err != nil {
		return nil, nil, err
	}
	if err := p.Prepare(r.Context(), d.Lab, assistant.Connection{MCPURL: a.MCPURL, Token: token}); err != nil {
		return nil, nil, scriptFailed(err)
	}
	return d, p, nil
}

func (a *api) chat(w http.ResponseWriter, r *http.Request) error {
	if a.AssistantURL == nil {
		return errChatOff
	}
	token, err := a.Tokens.Issue(r.Context(), r)
	if err != nil {
		return err
	}
	var unreachable error
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(a.AssistantURL)
			pr.Out.Header.Del("Cookie")
			pr.Out.Header.Set("Authorization", "Bearer "+token)
		},
		FlushInterval: -1,
		ErrorHandler:  func(_ http.ResponseWriter, _ *http.Request, err error) { unreachable = err },
	}
	proxy.ServeHTTP(w, r)
	if unreachable != nil {
		return domainerr.Unavailable("assistant_unreachable", fmt.Sprintf("reach the assistant: %v", unreachable)).AsFault()
	}
	return nil
}

type learner struct {
	Account string `json:"account"`
}

func (a *api) learner(w http.ResponseWriter, r *http.Request) error {
	claims, _ := auth.FromContext(r.Context())
	return httpx.WriteJSON(w, http.StatusOK, learner{Account: claims.AccountID})
}

type currentTask struct {
	Task     tasks.Task     `json:"task"`
	Question string         `json:"question"`
	Lab      sandbox.Status `json:"lab"`
}

func (a *api) currentTask(w http.ResponseWriter, r *http.Request) error {
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	if err := notDuringExam(d); err != nil {
		return err
	}
	t, ok := tasks.Find(a.Tasks, d.Practicing())
	if !ok {
		return errNoTask
	}
	return httpx.WriteJSON(w, http.StatusOK, currentTask{Task: t, Question: t.Question, Lab: d.Lab.Status()})
}

type inspectRequest struct {
	Args []string `json:"args"`
}

type inspection struct {
	Output    string `json:"output"`
	Failed    bool   `json:"failed,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

func (a *api) inspect(w http.ResponseWriter, r *http.Request) error {
	var req inspectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.Malformed(`the body must be JSON like {"args": ["get", "pods", "-n", "team-a"]}`)
	}
	if !readOnly(req.Args) {
		return errBadInspect
	}
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	if err := notDuringExam(d); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), inspectTimeout)
	defer cancel()
	out, err := d.Lab.Run(ctx, labagent.ControlPlaneNode, "inspect", []byte(kubectlScript(req.Args)))
	if _, ok := errors.AsType[*domainerr.Error](err); ok {
		return err
	}
	res := inspection{Output: out, Failed: err != nil}
	if res.Failed && out == "" {
		res.Output = err.Error()
	}
	if len(res.Output) > inspectLimit {
		res.Output, res.Truncated = res.Output[:inspectLimit], true
	}
	return httpx.WriteJSON(w, http.StatusOK, res)
}

func readOnly(args []string) bool {
	if len(args) == 0 || !slices.Contains(inspectVerbs, args[0]) {
		return false
	}
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		name, value, attached := flagParts(arg)
		if !slices.Contains(inspectFlags, name) {
			return false
		}
		if name != "-o" && name != "--output" {
			continue
		}
		if !attached && i+1 < len(args) {
			value = args[i+1]
		}
		if format, _, _ := strings.Cut(value, "="); strings.HasSuffix(format, "-file") {
			return false
		}
	}
	return true
}

func flagParts(arg string) (name, value string, attached bool) {
	if strings.HasPrefix(arg, "--") || len(arg) <= 2 {
		return strings.Cut(arg, "=")
	}
	return arg[:2], strings.TrimPrefix(arg[2:], "="), true
}

func kubectlScript(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
	}
	return "timeout 15 kubectl " + strings.Join(quoted, " ") + " 2>&1\n"
}
