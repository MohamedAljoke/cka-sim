package api

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

type taskView struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Domain      tasks.Domain `json:"domain"`
	DomainTitle string       `json:"domainTitle"`
	Topics      []string     `json:"topics"`
	Weight      int          `json:"weight"`
	Host        string       `json:"host"`
	Order       tasks.Order  `json:"order,omitempty"`
}

type taskDetailView struct {
	taskView
	Question string `json:"question"`
}

type solutionView struct {
	Solution string `json:"solution"`
}

// A setup that fails is still an answer: the page shows its exit code and the end of its output.
type setupView struct {
	Passed   bool   `json:"passed"`
	ExitCode int    `json:"exitCode"`
	Output   string `json:"output,omitempty"`
}

// A check that printed no PASS or FAIL line broke, so the page gets its exit code and output instead.
type gradeView struct {
	Earned   int         `json:"earned"`
	Total    int         `json:"total"`
	Checks   []checkView `json:"checks"`
	ExitCode int         `json:"exitCode,omitempty"`
	Output   string      `json:"output,omitempty"`
}

type checkView struct {
	Passed      bool   `json:"passed"`
	Points      int    `json:"points"`
	Description string `json:"description"`
}

func (a *api) taskRoutes(handle handleFunc) {
	handle("GET /api/tasks", a.getTasks)
	handle("GET /api/tasks/{id}", a.getTaskByID)
	handle("GET /api/tasks/{id}/solution", a.getTaskSolution)
	handle("POST /api/tasks/{id}/start", a.startTask)
	handle("POST /api/tasks/{id}/check", a.checkTask)
	handle("POST /api/tasks/{id}/tidy", a.tidyTask)
}

func (a *api) getTasks(w http.ResponseWriter, r *http.Request) error {
	all, err := a.tasks.List.Execute(r.Context(), filterOf(r.URL.Query()))
	switch {
	case errors.Is(err, tasks.ErrUnknownDomain), errors.Is(err, tasks.ErrMalformedTopic):
		return httpx.BadRequest(err)
	case errors.Is(err, tasks.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	views := make([]taskView, len(all))
	for i, t := range all {
		views[i] = newTaskView(t)
	}

	return httpx.WriteJSON(w, http.StatusOK, views)
}

func (a *api) getTaskByID(w http.ResponseWriter, r *http.Request) error {
	task, err := a.tasks.Get.Execute(r.Context(), r.PathValue("id"))
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		return httpx.NotFound(err)
	case errors.Is(err, tasks.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	return httpx.WriteJSON(w, http.StatusOK, taskDetailView{taskView: newTaskView(task), Question: task.Question})
}

func (a *api) getTaskSolution(w http.ResponseWriter, r *http.Request) error {
	if err := a.outsideAnExam(r); err != nil {
		return err
	}

	solution, err := a.tasks.Solution.Execute(r.Context(), r.PathValue("id"))
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		return httpx.NotFound(err)
	case errors.Is(err, tasks.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	return httpx.WriteJSON(w, http.StatusOK, solutionView{Solution: solution})
}

func (a *api) startTask(w http.ResponseWriter, r *http.Request) error {
	if err := a.outsideAnExam(r); err != nil {
		return err
	}

	result, err := a.tasks.Start.Execute(r.Context(), studentOf(r), r.PathValue("id"))
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		return httpx.NotFound(err)
	case errors.Is(err, labs.ErrNotReady):
		return httpx.Conflict(err)
	case errors.Is(err, tasks.ErrUnavailable), errors.Is(err, labs.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	view := setupView{Passed: result.Passed(), ExitCode: result.ExitCode, Output: result.Output}

	return httpx.WriteJSON(w, http.StatusOK, view)
}

func (a *api) checkTask(w http.ResponseWriter, r *http.Request) error {
	if err := a.outsideAnExam(r); err != nil {
		return err
	}

	checked, err := a.tasks.Check.Execute(r.Context(), studentOf(r), r.PathValue("id"))
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		return httpx.NotFound(err)
	case errors.Is(err, labs.ErrNotReady):
		return httpx.Conflict(err)
	case errors.Is(err, tasks.ErrUnavailable), errors.Is(err, labs.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	return httpx.WriteJSON(w, http.StatusOK, newGradeView(checked))
}

func (a *api) tidyTask(w http.ResponseWriter, r *http.Request) error {
	if err := a.outsideAnExam(r); err != nil {
		return err
	}

	err := a.tasks.Tidy.Execute(r.Context(), studentOf(r), r.PathValue("id"))
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		return httpx.NotFound(err)
	case errors.Is(err, labs.ErrNotReady):
		return httpx.Conflict(err)
	case errors.Is(err, tasks.ErrUnavailable), errors.Is(err, labs.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func newGradeView(checked usecases.Checked) gradeView {
	view := gradeViewOf(checked.Grade)

	if len(checked.Grade.Checks) == 0 {
		view.ExitCode, view.Output = checked.Script.ExitCode, checked.Script.Output
	}

	return view
}

func gradeViewOf(grade tasks.Grade) gradeView {
	view := gradeView{Earned: grade.Earned, Total: grade.Total, Checks: make([]checkView, len(grade.Checks))}

	for i, c := range grade.Checks {
		view.Checks[i] = checkView{Passed: c.Passed, Points: c.Points, Description: c.Description}
	}

	return view
}

func filterOf(query url.Values) tasks.Filter {
	filter := tasks.Filter{Topics: query["topic"], Search: query.Get("q")}
	for _, d := range query["domain"] {
		filter.Domains = append(filter.Domains, tasks.Domain(d))
	}

	return filter
}

func newTaskView(task tasks.Task) taskView {
	topics := task.Topics
	if topics == nil {
		topics = []string{}
	}

	return taskView{
		ID:          task.ID,
		Title:       task.Title,
		Domain:      task.Domain,
		DomainTitle: tasks.Titles[task.Domain],
		Topics:      topics,
		Weight:      task.Weight,
		Host:        task.Host,
		Order:       task.Order,
	}
}
