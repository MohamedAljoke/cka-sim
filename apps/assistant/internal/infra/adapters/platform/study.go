package platform

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	studyv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1/studyv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
)

// A start heals the cluster and runs the task's setup, which take minutes at most.
const startTimeout = 5 * time.Minute

// Study calls the platform's StudyService.
type Study struct {
	calls studyv1connect.StudyServiceClient
}

func NewStudy(url, token string) *Study {
	transport := connecthttp.NewTransport(&http.Client{Timeout: startTimeout}, url)

	return &Study{calls: studyv1connect.NewStudyServiceClient(connect.NewClient(transport, rpcauth.Bearer(token)))}
}

func (s *Study) ListTasks(ctx context.Context, student string, filter study.Filter) ([]study.TaskSummary, error) {
	res, err := s.calls.ListTasks(ctx, &studyv1.ListTasksRequest{
		Student: student, Domains: filter.Domains, Topics: filter.Topics, Search: filter.Search,
	})
	if err != nil {
		return nil, studyRefusal(err)
	}

	all := []study.TaskSummary{}
	for _, t := range res.GetTasks() {
		all = append(all, study.TaskSummary{
			ID: t.GetId(), Title: t.GetTitle(), Domain: t.GetDomain(), Topics: t.GetTopics(), Weight: int(t.GetWeight()),
		})
	}

	return all, nil
}

func (s *Study) StartTask(ctx context.Context, student, taskID string) (study.Started, error) {
	res, err := s.calls.StartTask(ctx, &studyv1.StartTaskRequest{Student: student, TaskId: taskID})
	if err != nil {
		return study.Started{}, studyRefusal(err)
	}

	task := res.GetTask()

	return study.Started{
		Task: study.Task{
			ID: task.GetId(), Title: task.GetTitle(), Domain: task.GetDomain(), Host: task.GetHost(),
			Question: task.GetQuestion(),
		},
		Ready:    res.GetReady(),
		ExitCode: int(res.GetExitCode()),
		Output:   res.GetOutput(),
	}, nil
}

func (s *Study) History(ctx context.Context, student string, taskIDs []string) ([]study.Entry, error) {
	res, err := s.calls.GetHistory(ctx, &studyv1.GetHistoryRequest{Student: student, TaskIds: taskIDs})
	if err != nil {
		return nil, studyRefusal(err)
	}

	entries := []study.Entry{}

	for _, entry := range res.GetEntries() {
		if note := entry.GetNote(); note != nil {
			entries = append(entries, study.Entry{
				Kind: study.NoteEntry, TaskID: note.GetTaskId(), At: note.GetAt().AsTime(), Text: note.GetText(),
			})

			continue
		}

		a := entry.GetAttempt()
		entries = append(entries, study.Entry{
			Kind: study.AttemptEntry, TaskID: a.GetTaskId(), At: a.GetAt().AsTime(), Earned: int(a.GetEarned()),
			Total: int(a.GetTotal()), Failed: a.GetFailed(),
		})
	}

	return entries, nil
}

func (s *Study) SaveNote(ctx context.Context, student, taskID, text string) error {
	_, err := s.calls.SaveNote(ctx, &studyv1.SaveNoteRequest{Student: student, TaskId: taskID, Text: text})
	if err != nil {
		return studyRefusal(err)
	}

	return nil
}

func studyRefusal(err error) error {
	return refusalAs(err, study.ErrRefused, study.ErrUnavailable)
}
