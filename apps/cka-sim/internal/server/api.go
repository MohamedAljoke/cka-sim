package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

type api struct {
	Practice
	setup sync.Mutex
}

func (a *api) listTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.Tasks)
}

func (a *api) startTask(w http.ResponseWriter, r *http.Request) {
	t, ok := tasks.Find(a.Tasks, r.PathValue("id"))
	if !ok {
		http.Error(w, fmt.Sprintf("no task %q", r.PathValue("id")), http.StatusNotFound)
		return
	}
	if !a.setup.TryLock() {
		http.Error(w, "a task is already being set up", http.StatusConflict)
		return
	}
	defer a.setup.Unlock()
	// Closing the tab mid-setup would otherwise leave the namespace half built.
	if err := tasks.Start(context.WithoutCancel(r.Context()), a.Files, a.Runner, t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"question": t.Question})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
