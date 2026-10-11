package server

import (
	"io/fs"
	"net/http"

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

// New serves the page's API. Every script and shell goes through lab to its current box.
func New(lab *sandbox.Lab, practice Practice) http.Handler {
	a := &api{Practice: practice, Lab: lab}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws/terminal", func(w http.ResponseWriter, r *http.Request) {
		labapi.Serve(w, r, lab)
	})
	mux.HandleFunc("GET /api/lab", a.getLab)
	mux.HandleFunc("POST /api/lab", a.startLab)
	mux.HandleFunc("DELETE /api/lab", a.endLab)
	mux.HandleFunc("GET /api/tasks", a.listTasks)
	mux.HandleFunc("GET /api/tasks/{id}/question", a.question)
	mux.HandleFunc("POST /api/tasks/{id}/start", a.startTask)
	mux.HandleFunc("POST /api/tasks/{id}/check", a.checkTask)
	mux.HandleFunc("GET /api/tasks/{id}/solution", a.solution)
	mux.HandleFunc("POST /api/tasks/{id}/tidy", a.tidyTask)
	if practice.Exam != nil {
		a.resumeExam()
		mux.HandleFunc("GET /api/exam", a.getExam)
		mux.HandleFunc("POST /api/exam", a.beginExam)
		mux.HandleFunc("PUT /api/exam/flags/{id}", a.flag)
		mux.HandleFunc("DELETE /api/exam/flags/{id}", a.flag)
		mux.HandleFunc("POST /api/exam/end", a.endExam)
		mux.HandleFunc("DELETE /api/exam", a.discardExam)
	}
	return http.NewCrossOriginProtection().Handler(mux)
}
