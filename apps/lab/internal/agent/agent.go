package agent

import (
	"context"
	"crypto/subtle"
	"io"
	"log"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/protocol"
)

const scriptLimit = 4 << 20 // 4 MiB

var errWrongToken = domainerr.Unauthorized("wrong_token", "wrong or missing token")

type Runner interface {
	Run(ctx context.Context, host, taskID string, script []byte) (string, error)
}

func Handler(token string, runner Runner, shells protocol.Opener) http.Handler {
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	authed := func(pattern string, h httpx.Handler) {
		mux.HandleFunc(pattern, problems.Handle(checkToken(token, h)))
	}
	authed("POST /run", func(w http.ResponseWriter, r *http.Request) error {
		return run(w, r, runner)
	})
	authed("GET /shell", func(w http.ResponseWriter, r *http.Request) error {
		command := r.URL.Query().Get("command")
		protocol.Serve(w, r, func(ctx context.Context, size protocol.Size) (protocol.Shell, error) {
			return shells.Open(ctx, command, size)
		})
		return nil
	})
	authed("POST /shells/end", func(w http.ResponseWriter, r *http.Request) error {
		if err := shells.EndAll(r.Context()); err != nil {
			return domainerr.Unavailable("end_shells_failed", err.Error()).AsFault()
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
	return mux
}

func checkToken(token string, next httpx.Handler) httpx.Handler {
	want := []byte("Bearer " + token)
	return func(w http.ResponseWriter, r *http.Request) error {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), want) != 1 {
			return errWrongToken
		}
		return next(w, r)
	}
}

// A failing script still answers 200: its error travels in the RunResult.
func run(w http.ResponseWriter, r *http.Request, runner Runner) error {
	host, taskID := r.URL.Query().Get("host"), r.URL.Query().Get("task")
	if host == "" || taskID == "" {
		return httpx.Malformed("host and task are required")
	}
	script, err := io.ReadAll(http.MaxBytesReader(w, r.Body, scriptLimit))
	if err != nil {
		return httpx.TooLarge(err.Error())
	}
	out, err := runner.Run(r.Context(), host, taskID, script)
	res := protocol.RunResult{Output: out}
	if err != nil {
		res.Error = err.Error()
	}
	return httpx.WriteJSON(w, http.StatusOK, res)
}
