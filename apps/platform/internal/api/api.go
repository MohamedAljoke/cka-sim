package api

import (
	"log"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

type api struct {
	listTasks   *usecases.ListTasks
	getTask     *usecases.GetTask
	getSolution *usecases.GetSolution
}

func New(listTasks *usecases.ListTasks, getTask *usecases.GetTask, getSolution *usecases.GetSolution) http.Handler {
	a := &api{listTasks: listTasks, getTask: getTask, getSolution: getSolution}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	handle := func(pattern string, h httpx.Handler) { mux.HandleFunc(pattern, problems.Handle(h)) }
	handle("GET /api/tasks", a.getTasks)
	handle("GET /api/tasks/{id}", a.getTaskByID)
	handle("GET /api/tasks/{id}/solution", a.getTaskSolution)
	return http.NewCrossOriginProtection().Handler(mux)
}
