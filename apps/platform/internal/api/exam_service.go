package api

import (
	"cmp"
	"context"
	"errors"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	examv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1/examv1connect"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

// examService lets the AI start an exam and read its result, never its questions. It isn't behind the gate's
// exam check: the exam use cases themselves refuse what can't happen while an exam runs.
type examService struct {
	examv1connect.UnimplementedExamServiceHandler
	gate
}

func (s *examService) StartExam(ctx context.Context, req *examv1.StartExamRequest) (*examv1.StartExamResponse, error) {
	if !s.knows(req.GetStudent()) {
		return nil, refuse(connect.CodeNotFound, errUnknownStudent, nil)
	}

	selection := tasks.Filter{Topics: req.GetTopics(), IDs: req.GetTaskIds()}
	for _, d := range req.GetDomains() {
		selection.Domains = append(selection.Domains, tasks.Domain(d))
	}

	exam, err := s.exams.Begin.Execute(ctx, req.GetStudent(), int(req.GetCount()), int(req.GetMinutes()), selection)
	switch {
	case errors.Is(err, tasks.ErrUnknownDomain), errors.Is(err, tasks.ErrMalformedTopic),
		errors.Is(err, tasks.ErrNotFound):
		// The message names the value refused, which came from the caller, so it can be shown as it is.
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error())
	case errors.Is(err, exams.ErrInvalid):
		return nil, refuse(connect.CodeInvalidArgument, exams.ErrInvalid, err)
	case errors.Is(err, exams.ErrNoneChosen):
		return nil, refuse(connect.CodeInvalidArgument, exams.ErrNoneChosen, err)
	case errors.Is(err, exams.ErrRunning):
		return nil, refuse(connect.CodeFailedPrecondition, exams.ErrRunning, err)
	case errors.Is(err, exams.ErrScoring):
		return nil, refuse(connect.CodeFailedPrecondition, exams.ErrScoring, err)
	case errors.Is(err, labs.ErrNotReady):
		return nil, refuse(connect.CodeFailedPrecondition, labs.ErrNotReady, err)
	case errors.Is(err, exams.ErrNoTasks):
		return nil, refuse(connect.CodeUnavailable, exams.ErrNoTasks, err)
	case errors.Is(err, exams.ErrUnavailable):
		return nil, refuse(connect.CodeUnavailable, exams.ErrUnavailable, err)
	case errors.Is(err, tasks.ErrUnavailable):
		return nil, refuse(connect.CodeUnavailable, tasks.ErrUnavailable, err)
	case errors.Is(err, labs.ErrUnavailable):
		return nil, refuse(connect.CodeUnavailable, labs.ErrUnavailable, err)
	case err != nil:
		return nil, err
	}

	res := &examv1.StartExamResponse{Minutes: int32(exam.Minutes), Tasks: []*examv1.ExamTask{}}
	for _, t := range exam.Tasks {
		res.Tasks = append(res.Tasks, &examv1.ExamTask{Id: t.ID, Title: t.Title})
	}

	return res, nil
}

func (s *examService) GetExamResult(ctx context.Context, req *examv1.GetExamResultRequest,
) (*examv1.GetExamResultResponse, error) {
	if !s.knows(req.GetStudent()) {
		return nil, refuse(connect.CodeNotFound, errUnknownStudent, nil)
	}

	exam, err := s.exams.Result.Execute(ctx, req.GetStudent())
	switch {
	case errors.Is(err, exams.ErrNoExam):
		return nil, refuse(connect.CodeFailedPrecondition, exams.ErrNoExam, err)
	case errors.Is(err, exams.ErrRunning):
		return nil, refuse(connect.CodeFailedPrecondition, exams.ErrRunning, err)
	case errors.Is(err, exams.ErrScoring):
		return nil, refuse(connect.CodeFailedPrecondition, exams.ErrScoring, err)
	case errors.Is(err, exams.ErrUnavailable):
		return nil, refuse(connect.CodeUnavailable, exams.ErrUnavailable, err)
	case err != nil:
		return nil, err
	}

	percent, passed := exams.Score(exam)
	res := &examv1.GetExamResultResponse{
		Tasks:   []*examv1.TaskResult{},
		Minutes: int32(exam.Minutes), Started: timestamppb.New(exam.Started), Ended: timestamppb.New(exam.Ended),
		Percent: int32(percent), Passed: passed, PassPercent: exams.PassPercent,
	}

	for _, task := range exam.Tasks {
		result := &examv1.TaskResult{
			Id: task.ID, Title: task.Title, Weight: int32(task.Weight), Flagged: task.Flagged,
			SetupFailed: task.Setup == exams.SetupFailed, Error: cmp.Or(task.SetupError, task.CheckError), Failed: []string{},
		}
		if task.Grade != nil {
			result.Earned, result.Total, result.Failed = int32(task.Grade.Earned), int32(task.Grade.Total), task.Grade.Failed()
		}

		res.Tasks = append(res.Tasks, result)
	}

	return res, nil
}
