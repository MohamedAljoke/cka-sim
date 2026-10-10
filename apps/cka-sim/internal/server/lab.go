package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
)

type labState struct {
	Lab sandbox.Status `json:"lab"`
	Now time.Time      `json:"now"`
}

func (a *api) getLab(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, labState{Lab: a.Lab.Status(), Now: time.Now()})
}

func (a *api) startLab(w http.ResponseWriter, r *http.Request) {
	a.Lab.Start()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(labState{Lab: a.Lab.Status(), Now: time.Now()})
}

func (a *api) endLab(w http.ResponseWriter, r *http.Request) {
	if a.Exam != nil && a.Exam.Running() {
		http.Error(w, "end the exam first", http.StatusConflict)
		return
	}
	if !a.lock(w, r) {
		return
	}
	defer a.busy.Unlock()
	if err := a.Lab.End(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
