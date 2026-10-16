package server

import (
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

type api struct {
	tasks []tasks.Task
}

type listedTask struct {
	tasks.Task
	DomainTitle string `json:"domainTitle"`
}

func (a *api) listTasks(w http.ResponseWriter, r *http.Request) error {
	listed := make([]listedTask, len(a.tasks))
	for i, t := range a.tasks {
		listed[i] = listedTask{Task: t, DomainTitle: tasks.Titles[t.Domain]}
	}
	return httpx.WriteJSON(w, http.StatusOK, listed)
}
