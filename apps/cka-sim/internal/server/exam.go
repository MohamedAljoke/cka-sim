package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/exam"
)

type examState struct {
	Exam  exam.Exam `json:"exam"`
	Now   time.Time `json:"now"`
	Score *score    `json:"score,omitempty"`
}

type score struct {
	Percent int  `json:"percent"`
	Passed  bool `json:"passed"`
}

type examRequest struct {
	Count   int `json:"count"`
	Minutes int `json:"minutes"`
}

func (a *api) getExam(w http.ResponseWriter, r *http.Request) {
	e, ok := a.Exam.Current()
	if !ok {
		http.Error(w, exam.ErrNoExam.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, a.state(e))
}

func (a *api) beginExam(w http.ResponseWriter, r *http.Request) {
	var req examRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "the body must be JSON like {\"count\": 16, \"minutes\": 120}", http.StatusBadRequest)
		return
	}
	if req.Count < 1 || req.Minutes < 1 || req.Minutes > 240 {
		http.Error(w, "count must be at least 1, and minutes between 1 and 240", http.StatusBadRequest)
		return
	}
	if !a.lock(w, r) {
		return
	}
	if err := a.Exam.Begin(req.Count, req.Minutes, a.busy.Unlock); err != nil {
		a.busy.Unlock()
		examError(w, err)
		return
	}
	e, _ := a.Exam.Current()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(a.state(e))
}

func (a *api) flag(w http.ResponseWriter, r *http.Request) {
	if err := a.Exam.Flag(r.PathValue("id"), r.Method == http.MethodPut); err != nil {
		examError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) endExam(w http.ResponseWriter, r *http.Request) {
	if !a.lock(w, r) {
		return
	}
	defer a.busy.Unlock()
	e, err := a.Exam.End(context.WithoutCancel(r.Context()))
	if err != nil {
		examError(w, err)
		return
	}
	writeJSON(w, a.state(e))
}

func (a *api) discardExam(w http.ResponseWriter, r *http.Request) {
	if !a.lock(w, r) {
		return
	}
	defer a.busy.Unlock()
	if err := a.Exam.Discard(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// resumeExam holds the script lock while setups a restart interrupted run again.
func (a *api) resumeExam() {
	a.busy.Lock()
	if !a.Exam.Resume(a.busy.Unlock) {
		a.busy.Unlock()
	}
}

func (a *api) state(e exam.Exam) examState {
	s := examState{Exam: e, Now: time.Now()}
	if e.Over() {
		percent, passed := exam.Score(e, a.Exam.Weights())
		s.Score = &score{Percent: percent, Passed: passed}
	}
	return s
}

func examError(w http.ResponseWriter, err error) {
	status := http.StatusConflict
	if errors.Is(err, exam.ErrNoExam) {
		status = http.StatusNotFound
	}
	http.Error(w, err.Error(), status)
}
