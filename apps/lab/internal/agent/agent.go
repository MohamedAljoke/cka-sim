// Package agent lets a server run task scripts and open shells on a lab that lives on another machine.
// The agent runs next to the lab's Docker and runs scripts and shells there; api.Client is the server's side.
package agent

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/lab/api"
)

const scriptLimit = 4 << 20 // 4 MiB

type Runner interface {
	Run(ctx context.Context, host, taskID string, script []byte) (string, error)
}

// Handler serves the agent. Scripts run as root on the cluster's nodes, so every route but the health
// check wants the token.
func Handler(token string, runner Runner, shells api.Opener) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.Handle("POST /run", authed(token, func(w http.ResponseWriter, r *http.Request) {
		run(w, r, runner)
	}))
	mux.Handle("GET /shell", authed(token, func(w http.ResponseWriter, r *http.Request) {
		api.Serve(w, r, shells)
	}))
	mux.Handle("POST /shells/end", authed(token, func(w http.ResponseWriter, r *http.Request) {
		if err := shells.EndAll(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	return mux
}

func authed(token string, next http.HandlerFunc) http.Handler {
	want := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), want) != 1 {
			http.Error(w, "wrong or missing token", http.StatusUnauthorized)
			return
		}
		next(w, r)
	})
}

// run's script stops when the server hangs up, since it runs on the request's context.
func run(w http.ResponseWriter, r *http.Request, runner Runner) {
	host, taskID := r.URL.Query().Get("host"), r.URL.Query().Get("task")
	if host == "" || taskID == "" {
		http.Error(w, "host and task are required", http.StatusBadRequest)
		return
	}
	script, err := io.ReadAll(http.MaxBytesReader(w, r.Body, scriptLimit))
	if err != nil {
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	out, err := runner.Run(r.Context(), host, taskID, script)
	res := api.RunResult{Output: out}
	if err != nil {
		res.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
