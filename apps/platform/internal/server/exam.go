package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/desk"
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
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	e, ok := d.Exam.Current()
	if !ok {
		return exam.ErrNoExam
	}
	return httpx.WriteJSON(w, http.StatusOK, state(d, e))
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
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	if err := lock(r, d); err != nil {
		return err
	}
	// The exam's setup waits for the lab, so starting an exam is enough to get one.
	d.Lab.Start()
	if err := d.Exam.Begin(pool, req.Count, req.Minutes, d.Unlock); err != nil {
		d.Unlock()
		return err
	}
	d.EndAssistants()
	e, _ := d.Exam.Current()
	return httpx.WriteJSON(w, http.StatusAccepted, state(d, e))
}

func (a *api) flag(w http.ResponseWriter, r *http.Request) error {
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	if err := d.Exam.Flag(r.PathValue("id"), r.Method == http.MethodPut); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// Answers at once; grading holds the script lock until every check is back.
func (a *api) endExam(w http.ResponseWriter, r *http.Request) error {
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	if err := lock(r, d); err != nil {
		return err
	}
	e, err := d.Exam.End(d.Unlock)
	if err != nil {
		d.Unlock()
		return err
	}
	return httpx.WriteJSON(w, http.StatusAccepted, state(d, e))
}

func (a *api) discardExam(w http.ResponseWriter, r *http.Request) error {
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	if err := lock(r, d); err != nil {
		return err
	}
	defer d.Unlock()
	if err := d.Exam.Discard(); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func state(d *desk.Desk, e exam.Exam) examState {
	s := examState{Exam: e, Now: time.Now()}
	if e.Graded() {
		percent, passed := exam.Score(e, d.Exam.Weights())
		s.Score = &score{Percent: percent, Passed: passed}
	}
	return s
}
