package api

import (
	"log/slog"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	chatusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats/usecases"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

type api struct {
	tasks usecases.UseCases
	labs  labusecases.UseCases
	exams examusecases.UseCases
	chats chatusecases.UseCases
	// token → student name
	students map[string]string
}

type handleFunc func(pattern string, h httpx.Handler)

func New(tasks usecases.UseCases, labs labusecases.UseCases, exams examusecases.UseCases,
	chats chatusecases.UseCases, students map[string]string,
) http.Handler {
	handlers := &api{tasks: tasks, labs: labs, exams: exams, chats: chats, students: students}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	}))
	mux := http.NewServeMux()
	handle := func(pattern string, h httpx.Handler) { mux.HandleFunc(pattern, problems.Handle(h)) }
	handlers.taskRoutes(handle)
	handlers.labRoutes(handle)
	handlers.terminalRoutes(handle)
	handlers.examRoutes(handle)
	handlers.studentRoutes(handle)
	handlers.chatRoutes(handle)

	return http.NewCrossOriginProtection().Handler(handlers.identify(mux, problems))
}
