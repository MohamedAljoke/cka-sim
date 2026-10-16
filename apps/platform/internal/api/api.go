package api

import (
	"log"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

type api struct {
	listTasks *usecases.ListTasks
}

func New(listTasks *usecases.ListTasks) http.Handler {
	a := &api{listTasks: listTasks}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	handle := func(pattern string, h httpx.Handler) { mux.HandleFunc(pattern, problems.Handle(h)) }
	handle("GET /api/tasks", a.getTasks)
	return http.NewCrossOriginProtection().Handler(mux)
}
