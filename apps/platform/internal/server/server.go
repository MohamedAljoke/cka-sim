package server

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"net/url"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/assistant"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/desk"
	labagent "github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

type Practice struct {
	Tasks        []tasks.Task
	Files        fs.FS
	Exams        bool
	Assistants   []assistant.Provider
	Tokens       assistant.Tokens
	MCPURL       string
	AssistantURL *url.URL
}

func New(desks *desk.Desks, practice Practice) http.Handler {
	a := &api{Practice: practice, desks: desks}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	handle := func(pattern string, h httpx.Handler) { mux.HandleFunc(pattern, problems.Handle(sessionOnly(h))) }
	assist := func(pattern string, h httpx.Handler) { mux.HandleFunc(pattern, problems.Handle(a.assistantToo(h))) }
	handle("GET /ws/terminal", func(w http.ResponseWriter, r *http.Request) error {
		d, err := a.desk(r)
		if err != nil {
			return err
		}
		labagent.Serve(w, r, func(ctx context.Context, size labagent.Size) (labagent.Shell, error) {
			return d.Open(ctx, "", size)
		})
		return nil
	})
	handle("GET /ws/assistant", a.serveAssistant)
	handle("POST /api/assistant", a.startAssistant)
	handle("POST /api/assistant/chat", a.chat)
	assist("GET /api/assistant/task", a.currentTask)
	assist("POST /api/assistant/inspect", a.inspect)
	assist("GET /api/assistant/learner", a.learner)
	handle("GET /api/lab", a.getLab)
	handle("POST /api/lab", a.startLab)
	handle("DELETE /api/lab", a.endLab)
	assist("GET /api/tasks", a.listTasks)
	assist("GET /api/tasks/{id}/question", a.question)
	handle("POST /api/tasks/{id}/start", a.startTask)
	assist("POST /api/tasks/{id}/check", a.checkTask)
	assist("GET /api/tasks/{id}/solution", a.solution)
	handle("POST /api/tasks/{id}/tidy", a.tidyTask)
	if practice.Exams {
		handle("GET /api/exam", a.getExam)
		handle("POST /api/exam", a.beginExam)
		handle("PUT /api/exam/flags/{id}", a.flag)
		handle("DELETE /api/exam/flags/{id}", a.flag)
		handle("POST /api/exam/end", a.endExam)
		handle("DELETE /api/exam", a.discardExam)
	}
	return http.NewCrossOriginProtection().Handler(mux)
}
