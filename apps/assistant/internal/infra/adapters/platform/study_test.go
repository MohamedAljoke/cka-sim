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

	studyv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1/studyv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
)

var noted = time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)

func TestListTasksSendsTheFilterForTheStudent(t *testing.T) {
	f := &fakeStudy{}

	got, err := NewStudy(serveStudy(t, f), token).ListTasks(context.Background(), "alice",
		study.Filter{Domains: []string{"workloads"}, Topics: []string{"poddisruptionbudgets"}, Search: "pdb"})

	if err != nil || len(got) != 1 || got[0].ID != "wl-pdb" || got[0].Weight != 4 || !slices.Equal(got[0].Topics, []string{"poddisruptionbudgets"}) {
		t.Errorf("got %+v, %v", got, err)
	}
	if f.list.GetStudent() != "alice" || f.list.GetDomains()[0] != "workloads" || f.list.GetSearch() != "pdb" {
		t.Errorf("sent %v", f.list)
	}
}

func TestStartTaskAnswersTheTaskAndHowItsSetupWent(t *testing.T) {
	got, err := NewStudy(serveStudy(t, &fakeStudy{}), token).StartTask(context.Background(), "alice", "wl-pdb")

	if err != nil || got.Task.Question != "Keep 2 running." || got.Task.Host != "cka-sim-control-plane" || got.Ready ||
		got.ExitCode != 1 || got.Output != "no image\n" {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestTheHistoryHasAttemptsAndNotes(t *testing.T) {
	got, err := NewStudy(serveStudy(t, &fakeStudy{}), token).History(context.Background(), "alice", []string{"wl-pdb"})

	if err != nil || len(got) != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got[0].Kind != study.NoteEntry || got[0].Text != "mixed up minAvailable" || !got[0].At.Equal(noted) {
		t.Errorf("note %+v", got[0])
	}
	if got[1].Kind != study.AttemptEntry || got[1].Earned != 2 || got[1].Total != 4 || !slices.Equal(got[1].Failed, []string{"selects the Pods"}) {
		t.Errorf("attempt %+v", got[1])
	}
}

func TestASaveNoteRefusalKeepsThePlatformsReason(t *testing.T) {
	f := &fakeStudy{err: connect.NewError(connect.CodeInvalidArgument, "note_too_long: a note is at most 500 characters")}

	err := NewStudy(serveStudy(t, f), token).SaveNote(context.Background(), "alice", "wl-pdb", "x")

	if !errors.Is(err, study.ErrRefused) || !strings.Contains(err.Error(), "at most 500 characters") {
		t.Errorf("got %v", err)
	}
}

func TestAStudyPlatformThatIsDownIsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	_, err := NewStudy(srv.URL, token).History(context.Background(), "alice", nil)

	if !errors.Is(err, study.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

type fakeStudy struct {
	studyv1connect.UnimplementedStudyServiceHandler
	err  error
	list *studyv1.ListTasksRequest
}

func (f *fakeStudy) ListTasks(_ context.Context, req *studyv1.ListTasksRequest) (*studyv1.ListTasksResponse, error) {
	f.list = req
	return &studyv1.ListTasksResponse{Tasks: []*studyv1.TaskSummary{
		{Id: "wl-pdb", Title: "PodDisruptionBudget", Domain: "workloads", Topics: []string{"poddisruptionbudgets"}, Weight: 4},
	}}, f.err
}

func (f *fakeStudy) StartTask(context.Context, *studyv1.StartTaskRequest) (*studyv1.StartTaskResponse, error) {
	return &studyv1.StartTaskResponse{
		Task:     &studyv1.Task{Id: "wl-pdb", Host: "cka-sim-control-plane", Question: "Keep 2 running."},
		ExitCode: 1, Output: "no image\n",
	}, f.err
}

func (f *fakeStudy) GetHistory(context.Context, *studyv1.GetHistoryRequest) (*studyv1.GetHistoryResponse, error) {
	return &studyv1.GetHistoryResponse{Entries: []*studyv1.Entry{
		{Kind: &studyv1.Entry_Note{Note: &studyv1.Note{TaskId: "wl-pdb", At: timestamppb.New(noted), Text: "mixed up minAvailable"}}},
		{Kind: &studyv1.Entry_Attempt{Attempt: &studyv1.Attempt{TaskId: "wl-pdb", At: timestamppb.New(noted), Earned: 2, Total: 4,
			Failed: []string{"selects the Pods"}}}},
	}}, f.err
}

func (f *fakeStudy) SaveNote(context.Context, *studyv1.SaveNoteRequest) (*studyv1.SaveNoteResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &studyv1.SaveNoteResponse{}, nil
}

func serveStudy(t *testing.T, f *fakeStudy) string {
	t.Helper()
	server := connect.NewServer(rpcauth.RequireBearer(token))
	studyv1connect.RegisterStudyServiceHandler(server, f)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}
