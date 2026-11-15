package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	studyv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1/studyv1connect"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

type study struct {
	studyv1connect.UnimplementedStudyServiceHandler
	gate
	tasks usecases.UseCases
}

func (s *study) ListTasks(ctx context.Context, req *studyv1.ListTasksRequest) (*studyv1.ListTasksResponse, error) {
	if err := s.outsideAnExam(ctx, req.GetStudent()); err != nil {
		return nil, err
	}

	filter := tasks.Filter{Topics: req.GetTopics(), Search: req.GetSearch()}
	for _, d := range req.GetDomains() {
		filter.Domains = append(filter.Domains, tasks.Domain(d))
	}

	all, err := s.tasks.List.Execute(ctx, filter)
	switch {
	case errors.Is(err, tasks.ErrUnknownDomain), errors.Is(err, tasks.ErrMalformedTopic):
		// The message names the value refused, which came from the caller, so it can be shown as it is.
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error())
	case errors.Is(err, tasks.ErrUnavailable):
		return nil, refuse(connect.CodeUnavailable, tasks.ErrUnavailable, err)
	case err != nil:
		return nil, err
	}

	res := &studyv1.ListTasksResponse{Tasks: []*studyv1.TaskSummary{}}
	for _, t := range all {
		res.Tasks = append(res.Tasks, &studyv1.TaskSummary{
			Id: t.ID, Title: t.Title, Domain: string(t.Domain), Topics: t.Topics, Weight: int32(t.Weight),
		})
	}

	return res, nil
}

func (s *study) StartTask(ctx context.Context, req *studyv1.StartTaskRequest) (*studyv1.StartTaskResponse, error) {
	if err := s.outsideAnExam(ctx, req.GetStudent()); err != nil {
		return nil, err
	}

	task, started, err := s.tasks.MoveTo.Execute(ctx, req.GetStudent(), req.GetTaskId())
	if err := refuseTaskOnLab(err); err != nil {
		return nil, err
	}

	res := &studyv1.StartTaskResponse{
		Task: &studyv1.Task{
			Id: task.ID, Title: task.Title, Domain: string(task.Domain), Host: task.Host, Question: task.Question,
		},
		Ready: started.Passed(),
	}
	if !started.Passed() {
		res.ExitCode, res.Output = int32(started.ExitCode), started.Output
	}

	return res, nil
}

func (s *study) GetHistory(ctx context.Context, req *studyv1.GetHistoryRequest) (*studyv1.GetHistoryResponse, error) {
	if err := s.outsideAnExam(ctx, req.GetStudent()); err != nil {
		return nil, err
	}

	entries, err := s.tasks.History.Execute(ctx, req.GetStudent(), req.GetTaskIds())
	switch {
	case errors.Is(err, tasks.ErrNoHistory):
		return nil, refuse(connect.CodeUnavailable, tasks.ErrNoHistory, err)
	case err != nil:
		return nil, err
	}

	res := &studyv1.GetHistoryResponse{Entries: []*studyv1.Entry{}}
	for _, e := range entries {
		res.Entries = append(res.Entries, entryOf(e))
	}

	return res, nil
}

func (s *study) SaveNote(ctx context.Context, req *studyv1.SaveNoteRequest) (*studyv1.SaveNoteResponse, error) {
	if err := s.outsideAnExam(ctx, req.GetStudent()); err != nil {
		return nil, err
	}

	err := s.tasks.Note.Execute(ctx, req.GetStudent(), req.GetTaskId(), req.GetText())
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		return nil, refuse(connect.CodeNotFound, tasks.ErrNotFound, err)
	case errors.Is(err, tasks.ErrEmptyNote):
		return nil, refuse(connect.CodeInvalidArgument, tasks.ErrEmptyNote, err)
	case errors.Is(err, tasks.ErrNoteTooLong):
		return nil, refuse(connect.CodeInvalidArgument, tasks.ErrNoteTooLong, err)
	case errors.Is(err, tasks.ErrUnavailable):
		return nil, refuse(connect.CodeUnavailable, tasks.ErrUnavailable, err)
	case errors.Is(err, tasks.ErrNoHistory):
		return nil, refuse(connect.CodeUnavailable, tasks.ErrNoHistory, err)
	case err != nil:
		return nil, err
	}

	return &studyv1.SaveNoteResponse{}, nil
}

func entryOf(entry tasks.Entry) *studyv1.Entry {
	if entry.Kind == tasks.NoteEntry {
		return &studyv1.Entry{Kind: &studyv1.Entry_Note{Note: &studyv1.Note{
			TaskId: entry.TaskID, At: timestamppb.New(entry.At), Text: entry.Text,
		}}}
	}

	return &studyv1.Entry{Kind: &studyv1.Entry_Attempt{Attempt: &studyv1.Attempt{
		TaskId: entry.TaskID, At: timestamppb.New(entry.At), Earned: int32(entry.Earned), Total: int32(entry.Total),
		Failed: entry.Failed,
	}}}
}
