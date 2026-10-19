package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1/tutorv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

var errUnknownStudent = domainerr.New("unknown_student", "no student has that name")

type tutor struct {
	tutorv1connect.UnimplementedTutorServiceHandler
	tasks usecases.UseCases
	labs  labusecases.UseCases
	exams examusecases.UseCases
	// token → student name
	students map[string]string
}

// WithTutor serves TutorService to whoever holds the token, beside the page's API.
func WithTutor(page http.Handler, token string, tasks usecases.UseCases, labs labusecases.UseCases,
	exams examusecases.UseCases, students map[string]string,
) http.Handler {
	server := connect.NewServer(hideFailures, rpcauth.RequireBearer(token))
	tutorv1connect.RegisterTutorServiceHandler(server,
		&tutor{tasks: tasks, labs: labs, exams: exams, students: students})
	mux := http.NewServeMux()
	mux.Handle("/", page)
	connecthttp.Mount(mux, server)

	return mux
}

func (t *tutor) GetCurrentTask(ctx context.Context, req *tutorv1.GetCurrentTaskRequest,
) (*tutorv1.GetCurrentTaskResponse, error) {
	if err := t.outsideAnExam(ctx, req.GetStudent()); err != nil {
		return nil, err
	}

	task, err := t.tasks.Current.Execute(ctx, req.GetStudent())
	switch {
	case errors.Is(err, tasks.ErrNoCurrentTask):
		return nil, refuse(connect.CodeNotFound, tasks.ErrNoCurrentTask, err)
	case errors.Is(err, tasks.ErrNotFound):
		return nil, refuse(connect.CodeNotFound, tasks.ErrNotFound, err)
	case errors.Is(err, tasks.ErrUnavailable):
		return nil, refuse(connect.CodeUnavailable, tasks.ErrUnavailable, err)
	case err != nil:
		return nil, err
	}

	return &tutorv1.GetCurrentTaskResponse{Task: &tutorv1.Task{
		Id:       task.ID,
		Title:    task.Title,
		Domain:   string(task.Domain),
		Host:     task.Host,
		Question: task.Question,
	}}, nil
}

func (t *tutor) CheckTask(ctx context.Context, req *tutorv1.CheckTaskRequest) (*tutorv1.CheckTaskResponse, error) {
	if err := t.outsideAnExam(ctx, req.GetStudent()); err != nil {
		return nil, err
	}

	checked, err := t.tasks.Check.Execute(ctx, req.GetStudent(), req.GetTaskId())
	if err := refuseTaskOnLab(err); err != nil {
		return nil, err
	}

	res := &tutorv1.CheckTaskResponse{Earned: int32(checked.Grade.Earned), Total: int32(checked.Grade.Total)}
	for _, c := range checked.Grade.Checks {
		res.Checks = append(res.Checks,
			&tutorv1.Check{Passed: c.Passed, Points: int32(c.Points), Description: c.Description})
	}

	if len(checked.Grade.Checks) == 0 {
		res.ExitCode, res.Output = int32(checked.Script.ExitCode), checked.Script.Output
	}

	return res, nil
}

func (t *tutor) InspectCluster(ctx context.Context, req *tutorv1.InspectClusterRequest,
) (*tutorv1.InspectClusterResponse, error) {
	if err := t.outsideAnExam(ctx, req.GetStudent()); err != nil {
		return nil, err
	}

	inspected, err := t.labs.Inspect.Execute(ctx, req.GetStudent(), req.GetArgs())
	switch {
	case errors.Is(err, labs.ErrNotReadOnly):
		// The message names the argument refused, which came from the caller, so it can be shown as it is.
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error())
	case errors.Is(err, labs.ErrNotReady):
		return nil, refuse(connect.CodeFailedPrecondition, labs.ErrNotReady, err)
	case errors.Is(err, labs.ErrUnavailable):
		return nil, refuse(connect.CodeUnavailable, labs.ErrUnavailable, err)
	case err != nil:
		return nil, err
	}

	return &tutorv1.InspectClusterResponse{
		Output:    inspected.Output,
		ExitCode:  int32(inspected.ExitCode),
		Truncated: inspected.Truncated,
	}, nil
}

func (t *tutor) GetSolution(ctx context.Context, req *tutorv1.GetSolutionRequest,
) (*tutorv1.GetSolutionResponse, error) {
	if err := t.outsideAnExam(ctx, req.GetStudent()); err != nil {
		return nil, err
	}

	solution, err := t.tasks.Solution.Execute(ctx, req.GetTaskId())
	if err := refuseTaskOnLab(err); err != nil {
		return nil, err
	}

	return &tutorv1.GetSolutionResponse{Solution: solution}, nil
}

// outsideAnExam lets a call through only for a known student who has no exam running.
func (t *tutor) outsideAnExam(ctx context.Context, student string) error {
	if !t.knows(student) {
		return refuse(connect.CodeNotFound, errUnknownStudent, nil)
	}

	exam, err := t.exams.Get.Execute(ctx, student)
	switch {
	case errors.Is(err, exams.ErrNoExam):
		return nil
	case errors.Is(err, exams.ErrUnavailable):
		return refuse(connect.CodeUnavailable, exams.ErrUnavailable, err)
	case err != nil:
		return err
	case !exam.Over():
		return refuse(connect.CodeFailedPrecondition, exams.ErrRunning, nil)
	case exam.Scoring():
		return refuse(connect.CodeFailedPrecondition, exams.ErrScoring, nil)
	}

	return nil
}

// With no students set, the only one is the local student.
func (t *tutor) knows(student string) bool {
	if len(t.students) == 0 {
		return student == localStudent
	}

	for _, name := range t.students {
		if name == student {
			return true
		}
	}

	return false
}

func refuseTaskOnLab(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, tasks.ErrNotFound):
		return refuse(connect.CodeNotFound, tasks.ErrNotFound, err)
	case errors.Is(err, labs.ErrNotReady):
		return refuse(connect.CodeFailedPrecondition, labs.ErrNotReady, err)
	case errors.Is(err, tasks.ErrUnavailable):
		return refuse(connect.CodeUnavailable, tasks.ErrUnavailable, err)
	case errors.Is(err, labs.ErrUnavailable):
		return refuse(connect.CodeUnavailable, labs.ErrUnavailable, err)
	}

	return err
}

func refuse(code connect.Code, sentinel *domainerr.Error, cause error) error {
	return connect.NewError(code, sentinel.Code+": "+sentinel.Message).WithCause(cause)
}

// hideFailures logs a failure no handler named and answers it as an internal error that never describes itself.
func hideFailures(next connect.ServerFunc) connect.ServerFunc {
	return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
		err := next(ctx, spec, stream)

		var named *connect.Error
		switch {
		case err == nil:
			return nil
		case !errors.As(err, &named):
			slog.Error("tutor call failed", "procedure", spec.Procedure, "err", err)
			return connect.NewError(connect.CodeInternal, "something went wrong; try again")
		case named.Code() == connect.CodeUnavailable:
			slog.Error("tutor call failed", "procedure", spec.Procedure, "err", errors.Unwrap(err))
		}

		return err
	}
}
