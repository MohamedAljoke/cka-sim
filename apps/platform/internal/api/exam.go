package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

const examBodyLimit = 1 << 10

type examRequest struct {
	Count   int `json:"count"`
	Minutes int `json:"minutes"`
}

// Now is the server's clock, so the page counts down against the deadline the server enforces.
type examView struct {
	Tasks    []examTaskView `json:"tasks"`
	Minutes  int            `json:"minutes"`
	Started  time.Time      `json:"started,omitzero"`
	Deadline time.Time      `json:"deadline,omitzero"`
	Ended    time.Time      `json:"ended,omitzero"`
	Scored   time.Time      `json:"scored,omitzero"`
	Now      time.Time      `json:"now"`
	Score    *scoreView     `json:"score,omitempty"`
}

type examTaskView struct {
	ID         string           `json:"id"`
	Weight     int              `json:"weight"`
	Setup      exams.SetupState `json:"setup"`
	SetupError string           `json:"setupError,omitempty"`
	Flagged    bool             `json:"flagged"`
	Grade      *gradeView       `json:"grade,omitempty"`
	CheckError string           `json:"checkError,omitempty"`
}

type scoreView struct {
	Percent int  `json:"percent"`
	Passed  bool `json:"passed"`
	Pass    int  `json:"pass"`
}

func (a *api) examRoutes(handle handleFunc) {
	handle("GET /api/exam", a.getExam)
	handle("POST /api/exam", a.beginExam)
	handle("DELETE /api/exam", a.discardExam)
	handle("POST /api/exam/end", a.endExam)
	handle("PUT /api/exam/tasks/{id}/flag", a.flagTask)
	handle("DELETE /api/exam/tasks/{id}/flag", a.flagTask)
}

func (a *api) getExam(w http.ResponseWriter, r *http.Request) error {
	exam, err := a.exams.Get.Execute(r.Context(), studentOf(r))
	switch {
	case errors.Is(err, exams.ErrNoExam):
		return httpx.NotFound(err)
	case errors.Is(err, exams.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	return httpx.WriteJSON(w, http.StatusOK, newExamView(exam))
}

func (a *api) beginExam(w http.ResponseWriter, r *http.Request) error {
	var req examRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, examBodyLimit)).Decode(&req); err != nil {
		return httpx.Malformed(`the body must be JSON like {"count": 16, "minutes": 120}`)
	}

	exam, err := a.exams.Begin.Execute(r.Context(), studentOf(r), req.Count, req.Minutes)
	switch {
	case errors.Is(err, exams.ErrInvalid):
		return httpx.BadRequest(err)
	case errors.Is(err, labs.ErrNotReady), errors.Is(err, exams.ErrRunning), errors.Is(err, exams.ErrScoring):
		return httpx.Conflict(err)
	case errors.Is(err, labs.ErrUnavailable), errors.Is(err, tasks.ErrUnavailable),
		errors.Is(err, exams.ErrUnavailable), errors.Is(err, exams.ErrNoTasks):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	return httpx.WriteJSON(w, http.StatusAccepted, newExamView(exam))
}

func (a *api) endExam(w http.ResponseWriter, r *http.Request) error {
	exam, err := a.exams.End.Execute(r.Context(), studentOf(r))
	switch {
	case errors.Is(err, exams.ErrNoExam):
		return httpx.NotFound(err)
	case errors.Is(err, exams.ErrEnded), errors.Is(err, exams.ErrSettingUp):
		return httpx.Conflict(err)
	case errors.Is(err, exams.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	return httpx.WriteJSON(w, http.StatusAccepted, newExamView(exam))
}

func (a *api) discardExam(w http.ResponseWriter, r *http.Request) error {
	err := a.exams.Discard.Execute(r.Context(), studentOf(r))
	switch {
	case errors.Is(err, exams.ErrNoExam):
		return httpx.NotFound(err)
	case errors.Is(err, exams.ErrRunning), errors.Is(err, exams.ErrScoring):
		return httpx.Conflict(err)
	case errors.Is(err, exams.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (a *api) flagTask(w http.ResponseWriter, r *http.Request) error {
	err := a.exams.Flag.Execute(r.Context(), studentOf(r), r.PathValue("id"), r.Method == http.MethodPut)
	switch {
	case errors.Is(err, exams.ErrNoExam), errors.Is(err, exams.ErrNotInExam):
		return httpx.NotFound(err)
	case errors.Is(err, exams.ErrEnded):
		return httpx.Conflict(err)
	case errors.Is(err, exams.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

// Practice runs scripts and shows solutions, so it waits until an exam is over and graded.
func (a *api) outsideAnExam(r *http.Request) error {
	exam, err := a.exams.Get.Execute(r.Context(), studentOf(r))
	switch {
	case errors.Is(err, exams.ErrNoExam):
		return nil
	case errors.Is(err, exams.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	case !exam.Over():
		return httpx.Conflict(exams.ErrRunning)
	case exam.Scoring():
		return httpx.Conflict(exams.ErrScoring)
	}

	return nil
}

func newExamView(exam exams.Exam) examView {
	view := examView{
		Tasks:    make([]examTaskView, len(exam.Tasks)),
		Minutes:  exam.Minutes,
		Started:  exam.Started,
		Deadline: exam.Deadline,
		Ended:    exam.Ended,
		Scored:   exam.Scored,
		Now:      time.Now(),
	}

	for i, t := range exam.Tasks {
		view.Tasks[i] = examTaskView{ID: t.ID, Weight: t.Weight, Setup: t.Setup, SetupError: t.SetupError, Flagged: t.Flagged}

		if t.Grade != nil {
			grade := gradeViewOf(*t.Grade)
			view.Tasks[i].Grade, view.Tasks[i].CheckError = &grade, t.CheckError
		}
	}

	if exam.Over() && !exam.Scoring() {
		percent, passed := exams.Score(exam)
		view.Score = &scoreView{Percent: percent, Passed: passed, Pass: exams.PassPercent}
	}

	return view
}
