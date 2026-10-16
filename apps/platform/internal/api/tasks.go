package api

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
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

func (a *api) taskRoutes(handle handleFunc) {
	handle("GET /api/tasks", a.getTasks)
	handle("GET /api/tasks/{id}", a.getTaskByID)
	handle("GET /api/tasks/{id}/solution", a.getTaskSolution)
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
	t, err := a.tasks.Get.Execute(r.Context(), r.PathValue("id"))
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		return httpx.NotFound(err)
	case errors.Is(err, tasks.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, taskDetailView{taskView: newTaskView(t), Question: t.Question})
}

func (a *api) getTaskSolution(w http.ResponseWriter, r *http.Request) error {
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

func filterOf(query url.Values) tasks.Filter {
	filter := tasks.Filter{Topics: query["topic"]}
	for _, d := range query["domain"] {
		filter.Domains = append(filter.Domains, tasks.Domain(d))
	}
	return filter
}

func newTaskView(t tasks.Task) taskView {
	topics := t.Topics
	if topics == nil {
		topics = []string{}
	}
	return taskView{
		ID:          t.ID,
		Title:       t.Title,
		Domain:      t.Domain,
		DomainTitle: tasks.Titles[t.Domain],
		Topics:      topics,
		Weight:      t.Weight,
		Host:        t.Host,
		Order:       t.Order,
	}
}
