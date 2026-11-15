package platform

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	examv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1/examv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam"
)

// The platform answers StartExam at once and sets the tasks up in the background.
const examTimeout = 30 * time.Second

type Exam struct {
	calls examv1connect.ExamServiceClient
}

func NewExam(url, token string) *Exam {
	transport := connecthttp.NewTransport(&http.Client{Timeout: examTimeout}, url)

	return &Exam{calls: examv1connect.NewExamServiceClient(connect.NewClient(transport, rpcauth.Bearer(token)))}
}

func (e *Exam) StartExam(ctx context.Context, student string, count, minutes int, selection exam.Selection,
) (exam.Exam, error) {
	res, err := e.calls.StartExam(ctx, &examv1.StartExamRequest{
		Student: student, Count: int32(count), Minutes: int32(minutes),
		Domains: selection.Domains, Topics: selection.Topics, TaskIds: selection.TaskIDs,
	})
	if err != nil {
		return exam.Exam{}, examRefusal(err)
	}

	started := exam.Exam{Minutes: int(res.GetMinutes()), Tasks: []exam.Task{}}
	for _, t := range res.GetTasks() {
		started.Tasks = append(started.Tasks, exam.Task{ID: t.GetId(), Title: t.GetTitle()})
	}

	return started, nil
}

func (e *Exam) ExamResult(ctx context.Context, student string) (exam.Result, error) {
	res, err := e.calls.GetExamResult(ctx, &examv1.GetExamResultRequest{Student: student})
	if err != nil {
		return exam.Result{}, examRefusal(err)
	}

	result := exam.Result{
		Tasks: []exam.TaskResult{}, Minutes: int(res.GetMinutes()), Started: res.GetStarted().AsTime(),
		Ended: res.GetEnded().AsTime(), Percent: int(res.GetPercent()), Passed: res.GetPassed(),
		PassPercent: int(res.GetPassPercent()),
	}
	for _, t := range res.GetTasks() {
		result.Tasks = append(result.Tasks, exam.TaskResult{
			ID: t.GetId(), Title: t.GetTitle(), Weight: int(t.GetWeight()), Earned: int(t.GetEarned()),
			Total: int(t.GetTotal()), Failed: t.GetFailed(), Flagged: t.GetFlagged(), SetupFailed: t.GetSetupFailed(),
			Error: t.GetError(),
		})
	}

	return result, nil
}

func examRefusal(err error) error {
	return refusalAs(err, exam.ErrRefused, exam.ErrUnavailable)
}
