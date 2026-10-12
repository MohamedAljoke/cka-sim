package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/auth"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/desk"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

var (
	errScriptRunning = domainerr.Conflict("script_running", "a script is already running")
	errDuringExam    = domainerr.Conflict("exam_running", "not during an exam")
)

type api struct {
	Practice
	desks *desk.Desks
}

func (a *api) desk(r *http.Request) (*desk.Desk, error) {
	claims, ok := auth.FromContext(r.Context())
	if !ok {
		return nil, auth.ErrNotLoggedIn
	}
	return a.desks.For(claims.AccountID)
}

type listedTask struct {
	tasks.Task
	DomainTitle string `json:"domainTitle"`
}

func (a *api) listTasks(w http.ResponseWriter, r *http.Request) error {
	listed := make([]listedTask, len(a.Tasks))
	for i, t := range a.Tasks {
		listed[i] = listedTask{Task: t, DomainTitle: tasks.Titles[t.Domain]}
	}
	return httpx.WriteJSON(w, http.StatusOK, listed)
}

func (a *api) question(w http.ResponseWriter, r *http.Request) error {
	t, err := a.findTask(r)
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, map[string]string{"question": t.Question})
}

func (a *api) startTask(w http.ResponseWriter, r *http.Request) error {
	d, t, err := a.practiceTask(r)
	if err != nil {
		return err
	}
	defer d.Unlock()
	d.Practice(t.ID)
	// A cancelled request stops the script; setup is idempotent, so the next start rebuilds it.
	ctx := r.Context()
	if err := tasks.Heal(ctx, a.Files, d.Lab, a.Tasks); err != nil {
		return scriptFailed(err)
	}
	if err := tasks.Start(ctx, a.Files, d.Lab, t); err != nil {
		return scriptFailed(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (a *api) checkTask(w http.ResponseWriter, r *http.Request) error {
	d, t, err := a.practiceTask(r)
	if err != nil {
		return err
	}
	defer d.Unlock()
	result, err := tasks.Check(r.Context(), a.Files, d.Lab, t)
	if err != nil {
		return scriptFailed(err)
	}
	return httpx.WriteJSON(w, http.StatusOK, result)
}

// The caller unlocks.
func (a *api) practiceTask(r *http.Request) (*desk.Desk, tasks.Task, error) {
	t, err := a.findTask(r)
	if err != nil {
		return nil, tasks.Task{}, err
	}
	d, err := a.desk(r)
	if err != nil {
		return nil, tasks.Task{}, err
	}
	if err := notDuringExam(d); err != nil {
		return nil, tasks.Task{}, err
	}
	if err := ensureLab(r, d); err != nil {
		return nil, tasks.Task{}, err
	}
	return d, t, lock(r, d)
}

func (a *api) tidyTask(w http.ResponseWriter, r *http.Request) error {
	t, err := a.findTask(r)
	if err != nil {
		return err
	}
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	if err := notDuringExam(d); err != nil {
		return err
	}
	d.Leave(t.ID)
	if d.Lab.Status().State != sandbox.Ready {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	if err := lock(r, d); err != nil {
		return err
	}
	go func() {
		defer d.Unlock()
		if err := tasks.Tidy(context.Background(), a.Files, d.Lab, []tasks.Task{t}); err != nil {
			log.Printf("tidy %s: %v", t.ID, err)
		}
	}()
	w.WriteHeader(http.StatusAccepted)
	return nil
}

func (a *api) solution(w http.ResponseWriter, r *http.Request) error {
	t, err := a.findTask(r)
	if err != nil {
		return err
	}
	d, err := a.desk(r)
	if err != nil {
		return err
	}
	if err := notDuringExam(d); err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, map[string]string{"explain": t.Explain})
}

// lockWait covers a script the page just cancelled, which takes a moment to stop, and the tidy
// after leaving a task, which may restart a kubelet that task broke.
var lockWait = 60 * time.Second

// Setup and check would otherwise race on the same namespace.
func lock(r *http.Request, d *desk.Desk) error {
	deadline := time.Now().Add(lockWait)
	for !d.TryLock() {
		if time.Now().After(deadline) || r.Context().Err() != nil {
			return errScriptRunning
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

func ensureLab(r *http.Request, d *desk.Desk) error {
	if err := d.Lab.Ensure(r.Context()); err != nil {
		return domainerr.Unavailable("lab_failed", "the lab failed to start: "+err.Error()).AsFault()
	}
	return nil
}

func notDuringExam(d *desk.Desk) error {
	if d.Exam != nil && d.Exam.Running() {
		return errDuringExam
	}
	return nil
}

func (a *api) findTask(r *http.Request) (tasks.Task, error) {
	id := r.PathValue("id")
	t, ok := tasks.Find(a.Tasks, id)
	if !ok {
		return tasks.Task{}, domainerr.NotFound("task_not_found", fmt.Sprintf("no task %q", id))
	}
	return t, nil
}

// Not an Internal error: the learner needs the script's reason to act on it.
func scriptFailed(err error) error {
	if _, ok := errors.AsType[*domainerr.Error](err); ok {
		return err
	}
	return domainerr.Unavailable("script_failed", err.Error()).AsFault()
}
