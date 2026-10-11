package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/exam"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

var errExamLimits = domainerr.Unprocessable("invalid_exam", "count must be at least 1, and minutes between 1 and 240")

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
	// Domains and Topics narrow the tasks drawn from; empty means the whole catalog.
	Domains []tasks.Domain `json:"domains"`
	Topics  []string       `json:"topics"`
}

func (a *api) getExam(w http.ResponseWriter, r *http.Request) error {
	e, ok := a.Exam.Current()
	if !ok {
		return exam.ErrNoExam
	}
	return httpx.WriteJSON(w, http.StatusOK, a.state(e))
}

func (a *api) beginExam(w http.ResponseWriter, r *http.Request) error {
	var req examRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.Malformed(`the body must be JSON like {"count": 16, "minutes": 120}`)
	}
	if req.Count < 1 || req.Minutes < 1 || req.Minutes > 240 {
		return errExamLimits
	}
	pool := tasks.Filter(a.Tasks, req.Domains, req.Topics)
	if len(pool) == 0 {
		return exam.ErrNoMatch
	}
	if err := a.lock(r); err != nil {
		return err
	}
	// The exam's setup waits for the lab, so starting an exam is enough to get one.
	a.Lab.Start()
	if err := a.Exam.Begin(pool, req.Count, req.Minutes, a.busy.Unlock); err != nil {
		a.busy.Unlock()
		return err
	}
	e, _ := a.Exam.Current()
	return httpx.WriteJSON(w, http.StatusAccepted, a.state(e))
}

func (a *api) flag(w http.ResponseWriter, r *http.Request) error {
	if err := a.Exam.Flag(r.PathValue("id"), r.Method == http.MethodPut); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// Answers at once; grading holds the script lock until every check is back.
func (a *api) endExam(w http.ResponseWriter, r *http.Request) error {
	if err := a.lock(r); err != nil {
		return err
	}
	e, err := a.Exam.End(a.busy.Unlock)
	if err != nil {
		a.busy.Unlock()
		return err
	}
	return httpx.WriteJSON(w, http.StatusAccepted, a.state(e))
}

func (a *api) discardExam(w http.ResponseWriter, r *http.Request) error {
	if err := a.lock(r); err != nil {
		return err
	}
	defer a.busy.Unlock()
	if err := a.Exam.Discard(); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (a *api) resumeExam() {
	a.busy.Lock()
	if !a.Exam.Resume(a.busy.Unlock) {
		a.busy.Unlock()
	}
}

func (a *api) state(e exam.Exam) examState {
	s := examState{Exam: e, Now: time.Now()}
	if e.Graded() {
		percent, passed := exam.Score(e, a.Exam.Weights())
		s.Score = &score{Percent: percent, Passed: passed}
	}
	return s
}
