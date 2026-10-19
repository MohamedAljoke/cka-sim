package api

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

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

const tutorToken = "unit-tutor-token-0123456789abcdef"

func TestTheCurrentTaskIsTheOneSetUpInThePage(t *testing.T) {
	h := newTutorAPI(&fakeRunner{}, nil)
	call(t, h, http.MethodPost, "/api/tasks/wl-scale/start")

	res, err := tutorClient(t, h, tutorToken).GetCurrentTask(context.Background(), &tutorv1.GetCurrentTaskRequest{Student: "local"})

	if err != nil || res.GetTask().GetId() != "wl-scale" || res.GetTask().GetQuestion() != "Scale it to 4." ||
		res.GetTask().GetDomain() != "workloads" {
		t.Errorf("got %v, %v", res, err)
	}
}

func TestNoCurrentTaskIsNotFound(t *testing.T) {
	_, err := tutorClient(t, newTutorAPI(&fakeRunner{}, nil), tutorToken).
		GetCurrentTask(context.Background(), &tutorv1.GetCurrentTaskRequest{Student: "local"})

	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "no_current_task") {
		t.Errorf("got %v", err)
	}
}

func TestACheckAnswersItsGrade(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "PASS 3 scaled\nFAIL 1 labelled\n"}}

	res, err := tutorClient(t, newTutorAPI(r, nil), tutorToken).
		CheckTask(context.Background(), &tutorv1.CheckTaskRequest{Student: "local", TaskId: "wl-scale"})

	if err != nil || res.GetEarned() != 3 || res.GetTotal() != 4 || len(res.GetChecks()) != 2 ||
		res.GetChecks()[1].GetDescription() != "labelled" || res.GetChecks()[1].GetPassed() {
		t.Errorf("got %v, %v", res, err)
	}
}

func TestInspectRunsAReadCommandOnTheControlPlane(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "web 1/1\n"}}

	res, err := tutorClient(t, newTutorAPI(r, nil), tutorToken).
		InspectCluster(context.Background(), &tutorv1.InspectClusterRequest{Student: "local", Args: []string{"get", "pods"}})

	if err != nil || res.GetOutput() != "web 1/1\n" || r.host != labs.ControlPlane || r.script != "timeout 15 kubectl 'get' 'pods'\n" {
		t.Errorf("got %v, %v; ran %q on %q", res, err, r.script, r.host)
	}
}

func TestACommandThatIsNotReadOnlyIsAnInvalidArgumentNamingIt(t *testing.T) {
	r := &fakeRunner{}

	_, err := tutorClient(t, newTutorAPI(r, nil), tutorToken).
		InspectCluster(context.Background(), &tutorv1.InspectClusterRequest{Student: "local", Args: []string{"delete", "pod", "web"}})

	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), `"delete"`) || r.script != "" {
		t.Errorf("got %v, ran %q", err, r.script)
	}
}

func TestTheSolutionComesWhenAsked(t *testing.T) {
	res, err := tutorClient(t, newTutorAPI(&fakeRunner{}, nil), tutorToken).
		GetSolution(context.Background(), &tutorv1.GetSolutionRequest{Student: "local", TaskId: "wl-scale"})

	if err != nil || res.GetSolution() != "Run kubectl scale." {
		t.Errorf("got %v, %v", res, err)
	}
}

func TestAnUnknownTaskIsNotFoundToTheTutor(t *testing.T) {
	_, err := tutorClient(t, newTutorAPI(&fakeRunner{}, nil), tutorToken).
		GetSolution(context.Background(), &tutorv1.GetSolutionRequest{Student: "local", TaskId: "no-such-task"})

	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "task_not_found") {
		t.Errorf("got %v", err)
	}
}

func TestALabThatIsNotReadyIsAFailedPrecondition(t *testing.T) {
	r := &fakeRunner{err: labs.ErrNotReady}

	_, err := tutorClient(t, newTutorAPI(r, nil), tutorToken).
		InspectCluster(context.Background(), &tutorv1.InspectClusterRequest{Student: "local", Args: []string{"get", "pods"}})

	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "lab_not_ready") {
		t.Errorf("got %v", err)
	}
}

func TestACheckOnALabThatDoesNotAnswerIsUnavailableAndHidesTheCause(t *testing.T) {
	r := &fakeRunner{err: errors.New("dial tcp 127.0.0.1:7171: connection refused")}

	_, err := tutorClient(t, newTutorAPI(r, nil), tutorToken).
		CheckTask(context.Background(), &tutorv1.CheckTaskRequest{Student: "local", TaskId: "wl-scale"})

	if connect.CodeOf(err) != connect.CodeUnavailable || strings.Contains(err.Error(), "7171") {
		t.Errorf("got %v", err)
	}
}

func TestTasksThatCannotLoadAreUnavailableAndHideTheCause(t *testing.T) {
	_, err := tutorClient(t, newTutorAPIWith(fakeRepository{err: errors.New("disk at /var/cka is gone")}, &fakeRunner{}, nil), tutorToken).
		GetSolution(context.Background(), &tutorv1.GetSolutionRequest{Student: "local", TaskId: "wl-scale"})

	if connect.CodeOf(err) != connect.CodeUnavailable || strings.Contains(err.Error(), "/var/cka") {
		t.Errorf("got %v", err)
	}
}

func TestAnUnknownStudentIsNotFound(t *testing.T) {
	h := newTutorAPI(&fakeRunner{}, map[string]string{"token-of-alice": "alice"})

	_, err := tutorClient(t, h, tutorToken).GetCurrentTask(context.Background(), &tutorv1.GetCurrentTaskRequest{Student: "local"})

	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "unknown_student") {
		t.Errorf("got %v", err)
	}
}

func TestAKnownStudentIsServed(t *testing.T) {
	h := newTutorAPI(&fakeRunner{}, map[string]string{"token-of-alice": "alice"})

	res, err := tutorClient(t, h, tutorToken).GetSolution(context.Background(), &tutorv1.GetSolutionRequest{Student: "alice", TaskId: "wl-scale"})

	if err != nil || res.GetSolution() == "" {
		t.Errorf("got %v, %v", res, err)
	}
}

func TestTheTutorWaitsWhileAnExamRuns(t *testing.T) {
	h := newTutorAPI(&fakeRunner{}, nil)
	send(t, h, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)

	_, err := tutorClient(t, h, tutorToken).GetSolution(context.Background(), &tutorv1.GetSolutionRequest{Student: "local", TaskId: "wl-scale"})

	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "exam_running") {
		t.Errorf("got %v", err)
	}
}

func TestAWrongTokenIsUnauthenticated(t *testing.T) {
	_, err := tutorClient(t, newTutorAPI(&fakeRunner{}, nil), "not-the-token").
		GetSolution(context.Background(), &tutorv1.GetSolutionRequest{Student: "local", TaskId: "wl-scale"})

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func TestThePageStillAnswersBesideTheTutor(t *testing.T) {
	status, body := call(t, newTutorAPI(&fakeRunner{}, nil), http.MethodGet, "/api/tasks/wl-scale")

	if status != http.StatusOK || !strings.Contains(body, `"question":"Scale it to 4."`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func newTutorAPI(r *fakeRunner, students map[string]string) http.Handler {
	return newTutorAPIWith(fakeRepository{tasks: []tasks.Task{wlScale}}, r, students)
}

func newTutorAPIWith(repo usecases.TaskRepository, r *fakeRunner, students map[string]string) http.Handler {
	taskCases := usecases.New(repo, r)
	labCases := labusecases.New(&fakeLabProvider{lab: labs.Lab{State: labs.Ready}}, r)
	examCases := examsFor(taskCases, labCases, nil)
	page := New(taskCases, labCases, examCases, students)

	return WithTutor(page, tutorToken, taskCases, labCases, examCases, students)
}

func tutorClient(t *testing.T, h http.Handler, token string) tutorv1connect.TutorServiceClient {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return tutorv1connect.NewTutorServiceClient(connect.NewClient(connecthttp.NewTransport(srv.Client(), srv.URL), rpcauth.Bearer(token)))
}
