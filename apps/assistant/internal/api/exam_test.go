package api

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam/usecases"
	studyusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
)

func TestStartExamSendsTheSelectionForTheStudent(t *testing.T) {
	p := &fakeExamPlatform{exam: exam.Exam{Minutes: 45, Tasks: []exam.Task{{ID: "nw-policy", Title: "Lock down egress"}}}}

	var got examOutput
	call(t, examSession(t, p), "start_exam",
		map[string]any{"count": 3, "minutes": 45, "topics": []string{"networkpolicies"}, "task_ids": []string{"nw-policy"}}, &got)

	if got.Minutes != 45 || len(got.Tasks) != 1 || got.Tasks[0] != (examTaskOutput{ID: "nw-policy", Title: "Lock down egress"}) {
		t.Errorf("got %+v", got)
	}
	if p.student != "alice" || p.count != 3 || p.minutes != 45 || !slices.Equal(p.selection.Topics, []string{"networkpolicies"}) ||
		!slices.Equal(p.selection.TaskIDs, []string{"nw-policy"}) {
		t.Errorf("asked %+v", p)
	}
}

func TestExamResultAnswersEachTasksGrade(t *testing.T) {
	ended := time.Date(2026, 11, 2, 11, 0, 0, 0, time.UTC)
	p := &fakeExamPlatform{result: exam.Result{Percent: 75, Passed: true, PassPercent: 66, Ended: ended, Tasks: []exam.TaskResult{
		{ID: "wl-scale", Earned: 3, Total: 4, Failed: []string{"labelled"}, Flagged: true},
	}}}

	var got resultOutput
	call(t, examSession(t, p), "exam_result", nil, &got)

	if got.Percent != 75 || !got.Passed || got.Ended != "2026-11-02T11:00:00Z" || len(got.Tasks) != 1 ||
		!slices.Equal(got.Tasks[0].Failed, []string{"labelled"}) || !got.Tasks[0].Flagged || p.student != "alice" {
		t.Errorf("got %+v", got)
	}
}

func TestAnExamRefusalReachesTheAIWithItsReason(t *testing.T) {
	p := &fakeExamPlatform{err: fmt.Errorf("%w: %s", exam.ErrRefused, "exam_running: an exam is running; end it first")}

	text := callFailing(t, examSession(t, p), "exam_result", nil)

	if !strings.Contains(text, "an exam is running") {
		t.Errorf("got %q", text)
	}
}

func TestAnExamPlatformThatIsDownSaysSoWithoutTheCause(t *testing.T) {
	p := &fakeExamPlatform{err: fmt.Errorf("%w: %w", exam.ErrUnavailable, errors.New("dial tcp 127.0.0.1:7070"))}

	text := callFailing(t, examSession(t, p), "start_exam", map[string]any{"count": 1, "minutes": 30})

	if !strings.Contains(text, "platform_unavailable") || strings.Contains(text, "7070") {
		t.Errorf("got %q", text)
	}
}

func TestThePagesTutorCannotStartAnExam(t *testing.T) {
	server := NewMCPServer(usecases.New(&fakePlatform{}), nil, "alice")
	serverSide, clientSide := mcp.NewInMemoryTransports()
	if _, err := server.Connect(context.Background(), serverSide, nil); err != nil {
		t.Fatal(err)
	}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(context.Background(), clientSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })

	res, err := session.ListTools(context.Background(), nil)

	if err != nil || slices.ContainsFunc(res.Tools, func(tool *mcp.Tool) bool { return tool.Name == "start_exam" || tool.Name == "exam_result" }) {
		t.Errorf("got %v, %v", res, err)
	}
}

func examSession(t *testing.T, p *fakeExamPlatform) *mcp.ClientSession {
	t.Helper()
	srv := httptest.NewServer(New(usecases.New(&fakePlatform{}), nil, studyusecases.New(&fakeStudyPlatform{}), examusecases.New(p),
		chatusecases.UseCases{}, "alice", token))
	t.Cleanup(srv.Close)
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).
		Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: srv.URL + "/mcp"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

type fakeExamPlatform struct {
	exam      exam.Exam
	result    exam.Result
	err       error
	student   string
	count     int
	minutes   int
	selection exam.Selection
}

func (p *fakeExamPlatform) StartExam(_ context.Context, student string, count, minutes int, selection exam.Selection) (exam.Exam, error) {
	p.student, p.count, p.minutes, p.selection = student, count, minutes, selection
	return p.exam, p.err
}

func (p *fakeExamPlatform) ExamResult(_ context.Context, student string) (exam.Result, error) {
	p.student = student
	return p.result, p.err
}
