package api

import (
	"net/http"

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

func (a *api) getTasks(w http.ResponseWriter, r *http.Request) error {
	all, err := a.listTasks.Execute(r.Context())
	if err != nil {
		return err
	}
	views := make([]taskView, len(all))
	for i, t := range all {
		views[i] = newTaskView(t)
	}
	return httpx.WriteJSON(w, http.StatusOK, views)
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
