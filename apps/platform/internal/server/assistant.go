package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/assistant"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/auth"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/desk"
	labagent "github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

var (
	errScoped     = domainerr.Forbidden("assistant_not_allowed", "an assistant token can't do this")
	errNoTask     = domainerr.NotFound("no_task", "the learner has no task open")
	errBadInspect = domainerr.Unprocessable("inspect_refused",
		"only read-only kubectl: get, describe, logs, events, top, explain, api-resources or version, without watching or following")
	errChatOff = domainerr.Unavailable("assistant_off", "no chat assistant is set up on this server")
	errBadChat = httpx.Malformed(fmt.Sprintf(`the body must be JSON like {"provider": "gemini", "messages": [{"role": "user", "text": "…"}]}`+
		` with 1 to %d messages, each up to %d bytes, alternating user and model and ending with user`, chatMessages, chatMessageMax))
)

const (
	inspectLimit   = 8 << 10
	inspectTimeout = 20 * time.Second
	chatMessages   = 40
	chatMessageMax = 8 << 10
	chatTimeout    = 90 * time.Second
)

var (
	inspectVerbs = []string{"get", "describe", "logs", "events", "top", "explain", "api-resources", "version"}
	// They never end on their own.
	inspectRefused = []string{"-w", "--watch", "--watch-only", "-f", "--follow"}
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

type chatRequest struct {
	Provider string        `json:"provider"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role llm.Role `json:"role"`
	Text string   `json:"text"`
}

func (a *api) chat(w http.ResponseWriter, r *http.Request) error {
	var req chatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, chatMessages*chatMessageMax*2)).Decode(&req); err != nil {
		return errBadChat
	}
	history, ok := chatHistory(req.Messages)
	if !ok {
		return errBadChat
	}
	c, ok := assistant.FindChat(a.Chats, req.Provider)
	if !ok && len(a.Chats) == 0 {
		return errChatOff
	}
	if !ok {
		return domainerr.Unprocessable("assistant_unknown", fmt.Sprintf("no assistant %q", req.Provider))
	}
	claims, _ := auth.FromContext(r.Context())
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	if err := notDuringExam(d); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), chatTimeout)
	defer cancel()
	token, err := a.Tokens.Issue(ctx, r)
	if err != nil {
		return err
	}
	tools, err := a.ChatTools.Open(ctx, token)
	if err != nil {
		return err
	}
	defer tools.Close()
	turn, err := c.Reply(ctx, claims.AccountID, tools, history)
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, turn)
}

func chatHistory(messages []chatMessage) ([]llm.Message, bool) {
	if len(messages) == 0 || len(messages) > chatMessages {
		return nil, false
	}
	history := make([]llm.Message, len(messages))
	for i, m := range messages {
		want := llm.RoleUser
		if i%2 == 1 {
			want = llm.RoleModel
		}
		if m.Role != want || m.Text == "" || len(m.Text) > chatMessageMax {
			return nil, false
		}
		history[i] = llm.Message{Role: m.Role, Text: m.Text}
	}
	return history, messages[len(messages)-1].Role == llm.RoleUser
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
	for _, arg := range args {
		name, _, _ := strings.Cut(arg, "=")
		if slices.Contains(inspectRefused, name) {
			return false
		}
	}
	return true
}

func kubectlScript(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
	}
	return "timeout 15 kubectl " + strings.Join(quoted, " ") + " 2>&1\n"
}
