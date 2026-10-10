package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

type api struct {
	Practice
	busy sync.Mutex
}

func (a *api) listTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.Tasks)
}

func (a *api) question(w http.ResponseWriter, r *http.Request) {
	t, ok := a.findTask(w, r)
	if !ok {
		return
	}
	writeJSON(w, map[string]string{"question": t.Question})
}

func (a *api) startTask(w http.ResponseWriter, r *http.Request) {
	t, ok := a.findTask(w, r)
	if !ok || a.duringExam(w) || !a.lock(w, r) {
		return
	}
	defer a.busy.Unlock()
	// A cancelled request stops the script; setup is idempotent, so the next start rebuilds it.
	ctx := r.Context()
	if err := tasks.Heal(ctx, a.Files, a.Runner, a.Tasks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tasks.Start(ctx, a.Files, a.Runner, t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) checkTask(w http.ResponseWriter, r *http.Request) {
	t, ok := a.findTask(w, r)
	if !ok || a.duringExam(w) || !a.lock(w, r) {
		return
	}
	defer a.busy.Unlock()
	result, err := tasks.Check(r.Context(), a.Files, a.Runner, t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

func (a *api) solution(w http.ResponseWriter, r *http.Request) {
	t, ok := a.findTask(w, r)
	if !ok || a.duringExam(w) {
		return
	}
	writeJSON(w, map[string]string{"explain": t.Explain})
}

// lockWait covers a script the page just cancelled: it takes a moment to stop and free the lock.
var lockWait = 10 * time.Second

// lock lets one script run at a time: setup and check would otherwise race on the same namespace.
func (a *api) lock(w http.ResponseWriter, r *http.Request) bool {
	deadline := time.Now().Add(lockWait)
	for !a.busy.TryLock() {
		if time.Now().After(deadline) || r.Context().Err() != nil {
			http.Error(w, "a script is already running", http.StatusConflict)
			return false
		}
		time.Sleep(50 * time.Millisecond)
	}
	return true
}

// duringExam keeps practice's checks and solutions away from a running exam.
func (a *api) duringExam(w http.ResponseWriter) bool {
	if a.Exam != nil && a.Exam.Running() {
		http.Error(w, "not during an exam", http.StatusConflict)
		return true
	}
	return false
}

func (a *api) findTask(w http.ResponseWriter, r *http.Request) (tasks.Task, bool) {
	id := r.PathValue("id")
	t, ok := tasks.Find(a.Tasks, id)
	if !ok {
		http.Error(w, fmt.Sprintf("no task %q", id), http.StatusNotFound)
	}
	return t, ok
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
