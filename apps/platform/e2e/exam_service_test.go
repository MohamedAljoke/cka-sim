package e2e

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	examv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/exam/v1/examv1connect"
	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	studyv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1"
	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestTheAIStartsAnExamOnTheTasksItChoseAndReviewsItsResult(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	url := ownServer(t, t.TempDir())
	client := examClient(url, assistantToken)

	started, err := client.StartExam(context.Background(), &examv1.StartExamRequest{
		Student: "local", Count: 5, Minutes: 30, TaskIds: []string{"wl-pdb", "nw-nodeport"},
	})
	if err != nil || started.GetMinutes() != 30 || len(started.GetTasks()) != 2 {
		t.Fatalf("start: %v, %v", started, err)
	}
	ids := []string{started.GetTasks()[0].GetId(), started.GetTasks()[1].GetId()}
	if slices.Sorted(slices.Values(ids))[0] != "nw-nodeport" || started.GetTasks()[0].GetTitle() == "" ||
		strings.Contains(started.String(), "never drop") {
		t.Errorf("start: %v", started)
	}
	e := waitForExam(t, url, func(e exam) bool { return e.Deadline != "" })
	if len(e.Tasks) != 2 {
		t.Fatalf("the page's exam: %+v", e)
	}

	_, err = client.GetExamResult(context.Background(), &examv1.GetExamResultRequest{Student: "local"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "exam_running") {
		t.Errorf("result while it runs: %v", err)
	}
	_, err = studyClient(url, assistantToken).StartTask(context.Background(), &studyv1.StartTaskRequest{Student: "local", TaskId: "wl-pdb"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "exam_running") {
		t.Errorf("study while it runs: %v", err)
	}
	_, err = tutorClientTo(url).GetSolution(context.Background(), &tutorv1.GetSolutionRequest{Student: "local", TaskId: "wl-pdb"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "exam_running") {
		t.Errorf("the tutor while it runs: %v", err)
	}

	requestTo(t, url, http.MethodPost, "/api/exam/end", nil)
	waitForExam(t, url, func(e exam) bool { return e.Score != nil })
	res, err := client.GetExamResult(context.Background(), &examv1.GetExamResultRequest{Student: "local"})

	if err != nil || len(res.GetTasks()) != 2 || res.GetPercent() != 67 || !res.GetPassed() || res.GetEnded().AsTime().IsZero() {
		t.Fatalf("result: %v, %v", res, err)
	}
	for _, task := range res.GetTasks() {
		if task.GetEarned() != 2 || task.GetTotal() != 3 || !slices.Equal(task.GetFailed(), []string{"not done yet"}) || task.GetTitle() == "" {
			t.Errorf("result of %s: %v", task.GetId(), task)
		}
	}
}

func TestStartExamWithoutASelectionDrawsFromTheWholeCatalog(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	url := ownServer(t, t.TempDir())

	res, err := examClient(url, assistantToken).StartExam(context.Background(),
		&examv1.StartExamRequest{Student: "local", Count: 3, Minutes: 30})

	if err != nil || len(res.GetTasks()) != 3 {
		t.Errorf("got %v, %v", res, err)
	}
	waitForExam(t, url, func(e exam) bool { return e.Deadline != "" })
}

func TestTheExamServiceRefusesWhatItCannotDo(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	url := ownServer(t, t.TempDir())
	client := examClient(url, assistantToken)

	for _, tc := range []struct {
		name string
		req  *examv1.StartExamRequest
		code connect.Code
		want string
	}{
		{"a count under 1", &examv1.StartExamRequest{Student: "local", Count: 0, Minutes: 30}, connect.CodeInvalidArgument, "invalid_exam"},
		{"too many minutes", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 241}, connect.CodeInvalidArgument, "invalid_exam"},
		{"a selection matching nothing", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30, Topics: []string{"no-such-topic"}},
			connect.CodeInvalidArgument, "no_task_selected"},
		{"an unknown task id", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30, TaskIds: []string{"wl-pdb", "no-such-task"}},
			connect.CodeInvalidArgument, `"no-such-task"`},
		{"a malformed topic", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30, Topics: []string{"Net Pol"}},
			connect.CodeInvalidArgument, "malformed_topic"},
		{"an unknown domain", &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30, Domains: []string{"security"}},
			connect.CodeInvalidArgument, "unknown_domain"},
		{"an unknown student", &examv1.StartExamRequest{Student: "mallory", Count: 2, Minutes: 30}, connect.CodeNotFound, "unknown_student"},
	} {
		if _, err := client.StartExam(context.Background(), tc.req); connect.CodeOf(err) != tc.code || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v", tc.name, err)
		}
	}

	_, err := client.GetExamResult(context.Background(), &examv1.GetExamResultRequest{Student: "local"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "no_exam") {
		t.Errorf("no exam yet: %v", err)
	}
	_, err = client.GetExamResult(context.Background(), &examv1.GetExamResultRequest{Student: "mallory"})
	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "unknown_student") {
		t.Errorf("an unknown student's result: %v", err)
	}
	_, err = examClient(url, agentToken).StartExam(context.Background(), &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30})
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("the agent's token: %v", err)
	}

	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})
	_, err = client.StartExam(context.Background(), &examv1.StartExamRequest{Student: "local", Count: 2, Minutes: 30})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "lab_not_ready") {
		t.Errorf("a lab that isn't ready: %v", err)
	}

	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	if _, err := client.StartExam(context.Background(), &examv1.StartExamRequest{Student: "local", Count: 1, Minutes: 30}); err != nil {
		t.Fatal(err)
	}
	_, err = client.StartExam(context.Background(), &examv1.StartExamRequest{Student: "local", Count: 1, Minutes: 30})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "exam_running") {
		t.Errorf("a second exam: %v", err)
	}
	waitForExam(t, url, func(e exam) bool { return e.Deadline != "" })
}

func examClient(url, token string) examv1connect.ExamServiceClient {
	return examv1connect.NewExamServiceClient(
		connect.NewClient(connecthttp.NewTransport(http.DefaultClient, url), rpcauth.Bearer(token)))
}
