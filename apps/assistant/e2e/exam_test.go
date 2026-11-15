package e2e

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	examv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1/examv1connect"
)

func TestStartExamStartsTheSelectionForTheConfiguredStudent(t *testing.T) {
	text := callTool(t, "start_exam", map[string]any{"count": 2, "minutes": 30, "topics": []string{"networkpolicies"}}, false)

	got := examPlatform.lastStart()
	if !strings.Contains(text, `{"id":"nw-policy","title":"Lock down egress"}`) || !strings.Contains(text, `"minutes":30`) ||
		got.GetStudent() != student || got.GetCount() != 2 || !slices.Equal(got.GetTopics(), []string{"networkpolicies"}) {
		t.Errorf("got %s; sent %v", text, got)
	}
}

func TestARefusedExamReachesTheAIWithThePlatformsReason(t *testing.T) {
	text := callTool(t, "start_exam", map[string]any{"count": 2, "minutes": 30, "task_ids": []string{"no-such-task"}}, true)

	if !strings.Contains(text, "no_task_selected") {
		t.Errorf("got %s", text)
	}
}

func TestExamResultAnswersWhatFailed(t *testing.T) {
	text := callTool(t, "exam_result", nil, false)

	if !strings.Contains(text, `"failed":["labelled"]`) || !strings.Contains(text, `"percent":75`) ||
		!strings.Contains(text, `"ended":"2026-11-02T11:00:00Z"`) {
		t.Errorf("got %s", text)
	}
}

func TestExamResultIsRefusedWhileTheExamRuns(t *testing.T) {
	examPlatform.setRunning(true)
	t.Cleanup(func() { examPlatform.setRunning(false) })

	text := callTool(t, "exam_result", nil, true)

	if !strings.Contains(text, "exam_running") || strings.Contains(text, "labelled") {
		t.Errorf("got %s", text)
	}
}

type fakeExam struct {
	examv1connect.UnimplementedExamServiceHandler
	mu      sync.Mutex
	start   *examv1.StartExamRequest
	running bool
}

func (f *fakeExam) lastStart() *examv1.StartExamRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.start
}

func (f *fakeExam) setRunning(running bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.running = running
}

func (f *fakeExam) StartExam(_ context.Context, req *examv1.StartExamRequest) (*examv1.StartExamResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if slices.Contains(req.GetTaskIds(), "no-such-task") {
		return nil, connect.NewError(connect.CodeInvalidArgument, "no_task_selected: no task matches that selection")
	}
	f.start = req
	return &examv1.StartExamResponse{Minutes: req.GetMinutes(), Tasks: []*examv1.ExamTask{{Id: "nw-policy", Title: "Lock down egress"}}}, nil
}

func (f *fakeExam) GetExamResult(context.Context, *examv1.GetExamResultRequest) (*examv1.GetExamResultResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.running {
		return nil, connect.NewError(connect.CodeFailedPrecondition, "exam_running: an exam is running; end it first")
	}
	return &examv1.GetExamResultResponse{
		Tasks:   []*examv1.TaskResult{{Id: "wl-scale", Title: "Scale", Weight: 4, Earned: 3, Total: 4, Failed: []string{"labelled"}}},
		Minutes: 30, Ended: timestamppb.New(time.Date(2026, 11, 2, 11, 0, 0, 0, time.UTC)), Percent: 75, Passed: true, PassPercent: 66,
	}, nil
}
