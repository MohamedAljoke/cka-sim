package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	examv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1/examv1connect"
	studyv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1"
	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	chatusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

func TestStartExamAnswersTheDrawnTasksWithoutTheirQuestions(t *testing.T) {
	res, err := examClient(t, newExamServiceAPI(&fakeRunner{}, labs.Ready, nil), tutorToken).StartExam(context.Background(),
		&examv1.StartExamRequest{Student: "local", Count: 1, Minutes: 30})

	if err != nil || res.GetMinutes() != 30 || len(res.GetTasks()) != 1 || res.GetTasks()[0].GetTitle() == "" {
		t.Fatalf("got %v, %v", res, err)
	}
	if strings.Contains(res.String(), "Scale it to 4") || strings.Contains(res.String(), "Find its IP") {
		t.Errorf("a question leaked: %v", res)
	}
}

func TestStartExamDrawsOnlyFromTheSelection(t *testing.T) {
	for name, req := range map[string]*examv1.StartExamRequest{
		"task ids": {Student: "local", Count: 4, Minutes: 30, TaskIds: []string{"nw-dns"}},
		"domains":  {Student: "local", Count: 4, Minutes: 30, Domains: []string{"networking"}},
		"topics":   {Student: "local", Count: 4, Minutes: 30, Topics: []string{"dns"}},
	} {
		res, err := examClient(t, newExamServiceAPI(&fakeRunner{}, labs.Ready, nil), tutorToken).StartExam(context.Background(), req)

		if err != nil || len(res.GetTasks()) != 1 || res.GetTasks()[0].GetId() != "nw-dns" || res.GetTasks()[0].GetTitle() != "Resolve a Service" {
			t.Errorf("%s: got %v, %v", name, res, err)
		}
	}
}

func TestStartExamRefusesWhatItCannotStart(t *testing.T) {
	for _, tc := range []struct {
		name string
		req  *examv1.StartExamRequest
		code connect.Code
		want string
	}{
		{"no tasks", &examv1.StartExamRequest{Student: "local", Count: 0, Minutes: 30}, connect.CodeInvalidArgument, "invalid_exam"},
		{"too long", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 241}, connect.CodeInvalidArgument, "invalid_exam"},
		{"no match", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30, Topics: []string{"no-such-topic"}},
			connect.CodeInvalidArgument, "no_task_selected"},
		{"unknown task", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30, TaskIds: []string{"nw-dns", "no-such-task"}},
			connect.CodeInvalidArgument, `task_not_found: no task has that id: "no-such-task"`},
		{"unknown domain", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30, Domains: []string{"security"}},
			connect.CodeInvalidArgument, `"security"`},
		{"malformed topic", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30, Topics: []string{"Net Pol"}},
			connect.CodeInvalidArgument, "malformed_topic"},
		{"unknown student", &examv1.StartExamRequest{Student: "mallory", Count: 2, Minutes: 30}, connect.CodeNotFound, "unknown_student"},
	} {
		_, err := examClient(t, newExamServiceAPI(&fakeRunner{}, labs.Ready, nil), tutorToken).StartExam(context.Background(), tc.req)

		if connect.CodeOf(err) != tc.code || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v", tc.name, err)
		}
	}
}

func TestStartExamWaitsForTheLab(t *testing.T) {
	_, err := examClient(t, newExamServiceAPI(&fakeRunner{}, labs.Starting, nil), tutorToken).StartExam(context.Background(),
		&examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30})

	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "lab_not_ready") {
		t.Errorf("got %v", err)
	}
}

func TestAnExamStartedByTheAIStopsTheTutorStudyAndItsOwnResult(t *testing.T) {
	h := newExamServiceAPI(&fakeRunner{}, labs.Ready, nil)
	client := examClient(t, h, tutorToken)
	if _, err := client.StartExam(context.Background(), &examv1.StartExamRequest{Student: "local", Count: 1, Minutes: 30}); err != nil {
		t.Fatal(err)
	}

	_, again := client.StartExam(context.Background(), &examv1.StartExamRequest{Student: "local", Count: 1, Minutes: 30})
	_, result := client.GetExamResult(context.Background(), &examv1.GetExamResultRequest{Student: "local"})
	_, list := studyClient(t, h, tutorToken).ListTasks(context.Background(), &studyv1.ListTasksRequest{Student: "local"})
	_, solution := tutorClient(t, h, tutorToken).GetSolution(context.Background(), &tutorv1.GetSolutionRequest{Student: "local", TaskId: "wl-scale"})

	for name, err := range map[string]error{"StartExam": again, "GetExamResult": result, "ListTasks": list, "GetSolution": solution} {
		if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "exam_running") {
			t.Errorf("%s: %v", name, err)
		}
	}
	if status, body := call(t, h, http.MethodGet, "/api/exam"); status != http.StatusOK || !strings.Contains(body, `"id":`) {
		t.Errorf("the page doesn't see it: status %d: %s", status, body)
	}
}

func TestNoExamHasNoResult(t *testing.T) {
	_, err := examClient(t, newExamServiceAPI(&fakeRunner{}, labs.Ready, nil), tutorToken).GetExamResult(context.Background(),
		&examv1.GetExamResultRequest{Student: "local"})

	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "no_exam") {
		t.Errorf("got %v", err)
	}
}

func TestTheResultOfAGradedExamNamesWhatFailed(t *testing.T) {
	h := newExamServiceAPI(&fakeRunner{result: labs.ScriptResult{Output: "PASS 3 scaled\nFAIL 1 labelled\n"}}, labs.Ready, nil)
	client := examClient(t, h, tutorToken)
	client.StartExam(context.Background(), &examv1.StartExamRequest{Student: "local", Count: 1, Minutes: 30, TaskIds: []string{"wl-scale"}})
	waitFor(t, h, `"deadline":`)
	call(t, h, http.MethodPost, "/api/exam/end")
	waitFor(t, h, `"score":`)

	res, err := client.GetExamResult(context.Background(), &examv1.GetExamResultRequest{Student: "local"})

	if err != nil || len(res.GetTasks()) != 1 || res.GetPercent() != 75 || !res.GetPassed() || res.GetPassPercent() != 66 ||
		res.GetEnded().AsTime().IsZero() {
		t.Fatalf("got %v, %v", res, err)
	}
	task := res.GetTasks()[0]
	if task.GetId() != "wl-scale" || task.GetTitle() != "Scale a Deployment" || task.GetEarned() != 3 || task.GetTotal() != 4 ||
		!slices.Equal(task.GetFailed(), []string{"labelled"}) || task.GetSetupFailed() {
		t.Errorf("got %v", task)
	}
}

func TestTheExamServiceRefusesAnUnknownStudentAndAWrongToken(t *testing.T) {
	h := newExamServiceAPI(&fakeRunner{}, labs.Ready, map[string]string{"token-of-alice": "alice"})

	_, err := examClient(t, h, tutorToken).GetExamResult(context.Background(), &examv1.GetExamResultRequest{Student: "local"})
	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "unknown_student") {
		t.Errorf("unknown student: %v", err)
	}

	_, err = examClient(t, h, "not-the-token").StartExam(context.Background(), &examv1.StartExamRequest{Student: "alice", Count: 1, Minutes: 30})
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("wrong token: %v", err)
	}
}

func newExamServiceAPI(r *fakeRunner, state labs.State, students map[string]string) http.Handler {
	taskCases := usecases.New(fakeRepository{tasks: []tasks.Task{wlScale, nwDNS}}, r, &fakeAttempts{}, time.Now)
	labCases := labusecases.New(&fakeLabProvider{lab: labs.Lab{State: state}}, r)
	examCases := examsFor(taskCases, labCases, nil)
	page := New(taskCases, labCases, examCases, chatusecases.UseCases{}, students)

	return WithTutor(page, tutorToken, taskCases, labCases, examCases, students)
}

func examClient(t *testing.T, h http.Handler, token string) examv1connect.ExamServiceClient {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return examv1connect.NewExamServiceClient(connect.NewClient(connecthttp.NewTransport(srv.Client(), srv.URL), rpcauth.Bearer(token)))
}
