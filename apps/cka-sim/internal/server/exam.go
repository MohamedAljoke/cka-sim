package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/exam"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
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
	// Domains and Topics narrow the tasks drawn from; empty means the whole catalog.
	Domains []tasks.Domain `json:"domains"`
	Topics  []string       `json:"topics"`
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
	pool := tasks.Filter(a.Tasks, req.Domains, req.Topics)
	if len(pool) == 0 {
		http.Error(w, "no task matches those domains and topics", http.StatusBadRequest)
		return
	}
	if !a.lock(w, r) {
		return
	}
	// The exam's setup waits for the lab, so starting an exam is enough to get one.
	a.Lab.Start()
	if err := a.Exam.Begin(pool, req.Count, req.Minutes, a.busy.Unlock); err != nil {
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

// endExam answers at once; grading holds the script lock until every check is back.
func (a *api) endExam(w http.ResponseWriter, r *http.Request) {
	if !a.lock(w, r) {
		return
	}
	e, err := a.Exam.End(a.busy.Unlock)
	if err != nil {
		a.busy.Unlock()
		examError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(a.state(e))
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

// resumeExam holds the script lock while setups or checks a restart interrupted run again.
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

func examError(w http.ResponseWriter, err error) {
	status := http.StatusConflict
	if errors.Is(err, exam.ErrNoExam) {
		status = http.StatusNotFound
	}
	http.Error(w, err.Error(), status)
}
