package server

import (
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

var errExamRunning = domainerr.Conflict("exam_running", "end the exam first")

type labState struct {
	Lab sandbox.Status `json:"lab"`
	Now time.Time      `json:"now"`
}

func (a *api) getLab(w http.ResponseWriter, r *http.Request) error {
	return httpx.WriteJSON(w, http.StatusOK, labState{Lab: a.Lab.Status(), Now: time.Now()})
}

func (a *api) startLab(w http.ResponseWriter, r *http.Request) error {
	a.Lab.Start()
	return httpx.WriteJSON(w, http.StatusAccepted, labState{Lab: a.Lab.Status(), Now: time.Now()})
}

func (a *api) endLab(w http.ResponseWriter, r *http.Request) error {
	if a.Exam != nil && a.Exam.Running() {
		return errExamRunning
	}
	if err := a.lock(r); err != nil {
		return err
	}
	defer a.busy.Unlock()
	if err := a.Lab.End(r.Context()); err != nil {
		return scriptFailed(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
