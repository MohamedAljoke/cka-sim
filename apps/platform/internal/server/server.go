package server

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	labapi "github.com/MohamedAljoke/cka-sim/apps/lab/api"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/exam"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

type Practice struct {
	Tasks []tasks.Task
	Files fs.FS
	Exam  *exam.Session
}

func New(lab *sandbox.Lab, practice Practice) http.Handler {
	a := &api{Practice: practice, Lab: lab}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	handle := func(pattern string, h httpx.Handler) { mux.HandleFunc(pattern, problems.Handle(h)) }
	mux.HandleFunc("GET /ws/terminal", func(w http.ResponseWriter, r *http.Request) {
		labapi.Serve(w, r, lab)
	})
	handle("GET /api/lab", a.getLab)
	handle("POST /api/lab", a.startLab)
	handle("DELETE /api/lab", a.endLab)
	handle("GET /api/tasks", a.listTasks)
	handle("GET /api/tasks/{id}/question", a.question)
	handle("POST /api/tasks/{id}/start", a.startTask)
	handle("POST /api/tasks/{id}/check", a.checkTask)
	handle("GET /api/tasks/{id}/solution", a.solution)
	handle("POST /api/tasks/{id}/tidy", a.tidyTask)
	if practice.Exam != nil {
		a.resumeExam()
		handle("GET /api/exam", a.getExam)
		handle("POST /api/exam", a.beginExam)
		handle("PUT /api/exam/flags/{id}", a.flag)
		handle("DELETE /api/exam/flags/{id}", a.flag)
		handle("POST /api/exam/end", a.endExam)
		handle("DELETE /api/exam", a.discardExam)
	}
	return http.NewCrossOriginProtection().Handler(mux)
}
