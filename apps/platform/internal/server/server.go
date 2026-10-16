package server

import (
	"log"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

func New(all []tasks.Task) http.Handler {
	a := &api{tasks: all}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	handle := func(pattern string, h httpx.Handler) { mux.HandleFunc(pattern, problems.Handle(h)) }
	handle("GET /api/tasks", a.listTasks)
	return http.NewCrossOriginProtection().Handler(mux)
}
