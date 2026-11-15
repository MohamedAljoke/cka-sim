package platform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"google.golang.org/protobuf/types/known/timestamppb"

	examv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1/examv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam"
)

var ended = time.Date(2026, 11, 2, 11, 0, 0, 0, time.UTC)

func TestStartExamSendsTheSelectionForTheStudent(t *testing.T) {
	f := &fakeExam{}

	got, err := NewExam(serveExam(t, f), token).StartExam(context.Background(), "alice", 3, 45,
		exam.Selection{Domains: []string{"networking"}, Topics: []string{"dns"}, TaskIDs: []string{"nw-dns"}})

	if err != nil || got.Minutes != 45 || len(got.Tasks) != 1 || got.Tasks[0] != (exam.Task{ID: "nw-dns", Title: "Resolve a Service"}) {
		t.Errorf("got %+v, %v", got, err)
	}
	if f.start.GetStudent() != "alice" || f.start.GetCount() != 3 || f.start.GetMinutes() != 45 ||
		!slices.Equal(f.start.GetDomains(), []string{"networking"}) || !slices.Equal(f.start.GetTopics(), []string{"dns"}) ||
		!slices.Equal(f.start.GetTaskIds(), []string{"nw-dns"}) {
		t.Errorf("sent %v", f.start)
	}
}

func TestTheResultHasEachTasksGrade(t *testing.T) {
	got, err := NewExam(serveExam(t, &fakeExam{}), token).ExamResult(context.Background(), "alice")

	if err != nil || got.Percent != 75 || !got.Passed || got.PassPercent != 66 || !got.Ended.Equal(ended) || len(got.Tasks) != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
	task := got.Tasks[0]
	if task.ID != "wl-scale" || task.Earned != 3 || task.Total != 4 || !slices.Equal(task.Failed, []string{"labelled"}) || !task.Flagged {
		t.Errorf("task %+v", task)
	}
}

func TestAnExamRefusalKeepsThePlatformsReason(t *testing.T) {
	f := &fakeExam{err: connect.NewError(connect.CodeFailedPrecondition, "exam_running: an exam is running; end it first")}

	_, err := NewExam(serveExam(t, f), token).ExamResult(context.Background(), "alice")

	if !errors.Is(err, exam.ErrRefused) || !strings.Contains(err.Error(), "an exam is running") {
		t.Errorf("got %v", err)
	}
}

func TestAnExamPlatformThatIsDownIsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	_, err := NewExam(srv.URL, token).StartExam(context.Background(), "alice", 1, 30, exam.Selection{})

	if !errors.Is(err, exam.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

type fakeExam struct {
	examv1connect.UnimplementedExamServiceHandler
	err   error
	start *examv1.StartExamRequest
}

func (f *fakeExam) StartExam(_ context.Context, req *examv1.StartExamRequest) (*examv1.StartExamResponse, error) {
	f.start = req
	if f.err != nil {
		return nil, f.err
	}
	return &examv1.StartExamResponse{Minutes: req.GetMinutes(), Tasks: []*examv1.ExamTask{{Id: "nw-dns", Title: "Resolve a Service"}}}, nil
}

func (f *fakeExam) GetExamResult(context.Context, *examv1.GetExamResultRequest) (*examv1.GetExamResultResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &examv1.GetExamResultResponse{
		Tasks:   []*examv1.TaskResult{{Id: "wl-scale", Title: "Scale", Weight: 4, Earned: 3, Total: 4, Failed: []string{"labelled"}, Flagged: true}},
		Minutes: 30, Ended: timestamppb.New(ended), Percent: 75, Passed: true, PassPercent: 66,
	}, nil
}

func serveExam(t *testing.T, f *fakeExam) string {
	t.Helper()
	server := connect.NewServer(rpcauth.RequireBearer(token))
	examv1connect.RegisterExamServiceHandler(server, f)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}
