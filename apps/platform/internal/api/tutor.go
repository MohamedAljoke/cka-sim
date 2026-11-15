package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1/examv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1/studyv1connect"
	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1/tutorv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

var errUnknownStudent = domainerr.New("unknown_student", "no student has that name")

type tutor struct {
	tutorv1connect.UnimplementedTutorServiceHandler
	gate
	tasks usecases.UseCases
	labs  labusecases.UseCases
}

// WithTutor serves TutorService, StudyService and ExamService to whoever holds the token, beside the page's API.
func WithTutor(page http.Handler, token string, tasks usecases.UseCases, labs labusecases.UseCases,
	exams examusecases.UseCases, students map[string]string,
) http.Handler {
	server := connect.NewServer(hideFailures, rpcauth.RequireBearer(token))
	known := gate{exams: exams, students: students}
	tutorv1connect.RegisterTutorServiceHandler(server, &tutor{gate: known, tasks: tasks, labs: labs})
	studyv1connect.RegisterStudyServiceHandler(server, &study{gate: known, tasks: tasks})
	examv1connect.RegisterExamServiceHandler(server, &examService{gate: known})
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

	current, err := t.tasks.Current.Execute(ctx, req.GetStudent())
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

	task := current.Task

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
