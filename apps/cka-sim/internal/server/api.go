package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

type api struct {
	Practice
	Lab  *sandbox.Lab
	busy sync.Mutex
}

type listedTask struct {
	tasks.Task
	DomainTitle string `json:"domainTitle"`
}

func (a *api) listTasks(w http.ResponseWriter, r *http.Request) {
	listed := make([]listedTask, len(a.Tasks))
	for i, t := range a.Tasks {
		listed[i] = listedTask{Task: t, DomainTitle: tasks.Titles[t.Domain]}
	}
	writeJSON(w, listed)
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
	if !ok || a.duringExam(w) || !a.ensureLab(w, r) || !a.lock(w, r) {
		return
	}
	defer a.busy.Unlock()
	// A cancelled request stops the script; setup is idempotent, so the next start rebuilds it.
	ctx := r.Context()
	if err := tasks.Heal(ctx, a.Files, a.Lab, a.Tasks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tasks.Start(ctx, a.Files, a.Lab, t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) checkTask(w http.ResponseWriter, r *http.Request) {
	t, ok := a.findTask(w, r)
	if !ok || a.duringExam(w) || !a.ensureLab(w, r) || !a.lock(w, r) {
		return
	}
	defer a.busy.Unlock()
	result, err := tasks.Check(r.Context(), a.Files, a.Lab, t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

// tidyTask cleans up after practice in the background, so the next setup finds nothing to delete.
func (a *api) tidyTask(w http.ResponseWriter, r *http.Request) {
	t, ok := a.findTask(w, r)
	if !ok || a.duringExam(w) {
		return
	}
	if a.Lab.Status().State != sandbox.Ready {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if !a.lock(w, r) {
		return
	}
	go func() {
		defer a.busy.Unlock()
		if err := tasks.Tidy(context.Background(), a.Files, a.Lab, []tasks.Task{t}); err != nil {
			log.Printf("tidy %s: %v", t.ID, err)
		}
	}()
	w.WriteHeader(http.StatusAccepted)
}

func (a *api) solution(w http.ResponseWriter, r *http.Request) {
	t, ok := a.findTask(w, r)
	if !ok || a.duringExam(w) {
		return
	}
	writeJSON(w, map[string]string{"explain": t.Explain})
}

// lockWait covers a script the page just cancelled, which takes a moment to stop, and the tidy
// after leaving a task, which may restart a kubelet that task broke.
var lockWait = 60 * time.Second

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

// ensureLab starts the lab if there is none, so practising a task is enough to get one.
func (a *api) ensureLab(w http.ResponseWriter, r *http.Request) bool {
	if err := a.Lab.Ensure(r.Context()); err != nil {
		http.Error(w, "the lab failed to start: "+err.Error(), http.StatusServiceUnavailable)
		return false
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
