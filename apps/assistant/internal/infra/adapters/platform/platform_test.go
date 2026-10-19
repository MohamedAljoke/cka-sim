package platform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1/tutorv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

const token = "unit-assistant-token-0123456789abcdef"

func TestTheCurrentTaskComesFromThePlatformForTheStudent(t *testing.T) {
	f := &fakeTutor{}

	task, err := New(serve(t, f), token).CurrentTask(context.Background(), "alice")

	if err != nil || task.ID != "wl-scale" || task.Question != "Scale it to 4." || f.student != "alice" {
		t.Errorf("got %+v, %v; asked for %q", task, err, f.student)
	}
}

func TestACheckAnswersEveryCheck(t *testing.T) {
	grade, err := New(serve(t, &fakeTutor{}), token).Check(context.Background(), "local", "wl-scale")

	if err != nil || grade.Earned != 3 || grade.Total != 4 || len(grade.Checks) != 2 || grade.Checks[1].Passed ||
		grade.Checks[1].Description != "labelled" {
		t.Errorf("got %+v, %v", grade, err)
	}
}

func TestAnInspectionSendsTheArgsAsTheyAre(t *testing.T) {
	f := &fakeTutor{}

	got, err := New(serve(t, f), token).Inspect(context.Background(), "local", []string{"get", "pods", "-A"})

	if err != nil || got.Output != "web 1/1\n" || !got.Truncated || strings.Join(f.args, " ") != "get pods -A" {
		t.Errorf("got %+v, %v; sent %q", got, err, f.args)
	}
}

func TestARefusalKeepsThePlatformsReason(t *testing.T) {
	for _, code := range []connect.Code{connect.CodeNotFound, connect.CodeInvalidArgument, connect.CodeFailedPrecondition} {
		f := &fakeTutor{err: connect.NewError(code, `not_read_only: "delete" isn't allowed`)}

		_, err := New(serve(t, f), token).Inspect(context.Background(), "local", []string{"delete"})

		if !errors.Is(err, tutor.ErrRefused) || !strings.Contains(err.Error(), `"delete" isn't allowed`) {
			t.Errorf("%v: got %v", code, err)
		}
	}
}

func TestAPlatformThatIsDownIsUnavailable(t *testing.T) {
	_, err := New("http://127.0.0.1:1", token).Solution(context.Background(), "local", "wl-scale")

	if !errors.Is(err, tutor.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestAWrongTokenIsNeitherARefusalNorUnavailable(t *testing.T) {
	_, err := New(serve(t, &fakeTutor{}), "not-the-token").Solution(context.Background(), "local", "wl-scale")

	if err == nil || errors.Is(err, tutor.ErrRefused) || errors.Is(err, tutor.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

type fakeTutor struct {
	tutorv1connect.UnimplementedTutorServiceHandler
	err     error
	student string
	args    []string
}

func (f *fakeTutor) GetCurrentTask(_ context.Context, req *tutorv1.GetCurrentTaskRequest,
) (*tutorv1.GetCurrentTaskResponse, error) {
	f.student = req.GetStudent()
	return &tutorv1.GetCurrentTaskResponse{Task: &tutorv1.Task{Id: "wl-scale", Question: "Scale it to 4."}}, f.err
}

func (f *fakeTutor) CheckTask(context.Context, *tutorv1.CheckTaskRequest) (*tutorv1.CheckTaskResponse, error) {
	return &tutorv1.CheckTaskResponse{Earned: 3, Total: 4, Checks: []*tutorv1.Check{
		{Passed: true, Points: 3, Description: "scaled"},
		{Passed: false, Points: 1, Description: "labelled"},
	}}, f.err
}

func (f *fakeTutor) InspectCluster(_ context.Context, req *tutorv1.InspectClusterRequest,
) (*tutorv1.InspectClusterResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.args = req.GetArgs()
	return &tutorv1.InspectClusterResponse{Output: "web 1/1\n", Truncated: true}, nil
}

func (f *fakeTutor) GetSolution(context.Context, *tutorv1.GetSolutionRequest) (*tutorv1.GetSolutionResponse, error) {
	return &tutorv1.GetSolutionResponse{Solution: "Run kubectl scale."}, f.err
}

func serve(t *testing.T, f *fakeTutor) string {
	t.Helper()
	server := connect.NewServer(rpcauth.RequireBearer(token))
	tutorv1connect.RegisterTutorServiceHandler(server, f)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}
