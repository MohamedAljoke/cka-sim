package api

import (
	"log"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

type api struct {
	tasks usecases.UseCases
	labs  labusecases.UseCases
}

type handleFunc func(pattern string, h httpx.Handler)

func New(tasks usecases.UseCases, labs labusecases.UseCases) http.Handler {
	a := &api{tasks: tasks, labs: labs}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	handle := func(pattern string, h httpx.Handler) { mux.HandleFunc(pattern, problems.Handle(h)) }
	a.taskRoutes(handle)
	a.labRoutes(handle)
	return http.NewCrossOriginProtection().Handler(mux)
}
