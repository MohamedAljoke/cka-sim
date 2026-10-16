package api

import (
	"log"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

type api struct {
	listTasks   *usecases.ListTasks
	getTask     *usecases.GetTask
	getSolution *usecases.GetSolution
	getLab      *labusecases.GetLab
	startLab    *labusecases.StartLab
	stopLab     *labusecases.StopLab
}

func New(
	listTasks *usecases.ListTasks, getTask *usecases.GetTask, getSolution *usecases.GetSolution,
	getLab *labusecases.GetLab, startLab *labusecases.StartLab, stopLab *labusecases.StopLab,
) http.Handler {
	a := &api{listTasks: listTasks, getTask: getTask, getSolution: getSolution, getLab: getLab, startLab: startLab, stopLab: stopLab}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	handle := func(pattern string, h httpx.Handler) { mux.HandleFunc(pattern, problems.Handle(h)) }
	handle("GET /api/tasks", a.getTasks)
	handle("GET /api/tasks/{id}", a.getTaskByID)
	handle("GET /api/tasks/{id}/solution", a.getTaskSolution)
	handle("GET /api/lab", a.getLabStatus)
	handle("POST /api/lab", a.postLab)
	handle("DELETE /api/lab", a.deleteLab)
	return http.NewCrossOriginProtection().Handler(mux)
}
