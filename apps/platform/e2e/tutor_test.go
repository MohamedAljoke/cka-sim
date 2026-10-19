package e2e

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1/tutorv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestTheTutorFollowsTheTaskSetUpInThePage(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	tutor := tutorClient(assistantToken)

	request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/start", nil)
	res, err := tutor.GetCurrentTask(context.Background(), &tutorv1.GetCurrentTaskRequest{Student: "local"})
	if err != nil || res.GetTask().GetId() != "ar-aggregate-view" || res.GetTask().GetQuestion() == "" {
		t.Fatalf("after setting up: %v, %v", res, err)
	}

	request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/tidy", nil)
	_, err = tutor.GetCurrentTask(context.Background(), &tutorv1.GetCurrentTaskRequest{Student: "local"})
	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "no_current_task") {
		t.Errorf("after leaving: %v", err)
	}
}

func TestTheTutorRunsTheRealCheck(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	res, err := tutorClient(assistantToken).CheckTask(context.Background(),
		&tutorv1.CheckTaskRequest{Student: "local", TaskId: "ar-aggregate-view"})

	if err != nil || res.GetEarned() != 2 || res.GetTotal() != 3 || len(res.GetChecks()) != 2 ||
		res.GetChecks()[0].GetDescription() != "checked on cka-sim-control-plane" {
		t.Errorf("got %v, %v", res, err)
	}
}

func TestTheTutorInspectsTheClusterFromTheControlPlane(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	ran := agent.scriptsFrom()

	res, err := tutorClient(assistantToken).InspectCluster(context.Background(),
		&tutorv1.InspectClusterRequest{Student: "local", Args: []string{"get", "pods", "-A"}})

	if err != nil || res.GetExitCode() != 2 || res.GetOutput() != "not a catalog script\n" {
		t.Errorf("got %v, %v", res, err)
	}
	if got := ran(); len(got) != 1 || got[0] != "timeout 15 kubectl 'get' 'pods' '-A'\n" {
		t.Errorf("the agent ran %q", got)
	}
}

func TestTheTutorCannotChangeTheCluster(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	ran := agent.scriptsFrom()

	for _, args := range [][]string{
		{"delete", "pod", "web"},
		{"get", "pods;", "rm", "-rf", "/"},
		{"get", "pods", "--kubeconfig=/x"},
	} {
		_, err := tutorClient(assistantToken).InspectCluster(context.Background(),
			&tutorv1.InspectClusterRequest{Student: "local", Args: args})

		if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "not_read_only") {
			t.Errorf("%q: %v", args, err)
		}
	}
	if got := ran(); len(got) != 0 {
		t.Errorf("the agent ran %q", got)
	}
}

func TestTheTutorWaitsForTheLab(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})

	_, err := tutorClient(assistantToken).InspectCluster(context.Background(),
		&tutorv1.InspectClusterRequest{Student: "local", Args: []string{"get", "pods"}})

	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "lab_not_ready") {
		t.Errorf("got %v", err)
	}
}

func TestTheTutorGivesTheSolutionWhenAsked(t *testing.T) {
	res, err := tutorClient(assistantToken).GetSolution(context.Background(),
		&tutorv1.GetSolutionRequest{Student: "local", TaskId: "ar-aggregate-view"})

	if err != nil || res.GetSolution() == "" {
		t.Errorf("got %v, %v", res, err)
	}
}

func TestTheTutorRefusesAnUnknownTaskAndStudent(t *testing.T) {
	_, err := tutorClient(assistantToken).GetSolution(context.Background(),
		&tutorv1.GetSolutionRequest{Student: "local", TaskId: "no-such-task"})
	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "task_not_found") {
		t.Errorf("unknown task: %v", err)
	}

	_, err = tutorClient(assistantToken).GetSolution(context.Background(),
		&tutorv1.GetSolutionRequest{Student: "mallory", TaskId: "ar-aggregate-view"})
	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "unknown_student") {
		t.Errorf("unknown student: %v", err)
	}
}

func TestTheTutorNeedsTheAssistantsToken(t *testing.T) {
	for _, token := range []string{"", agentToken} {
		_, err := tutorClient(token).GetSolution(context.Background(),
			&tutorv1.GetSolutionRequest{Student: "local", TaskId: "ar-aggregate-view"})

		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("token %q: %v", token, err)
		}
	}
}

func tutorClient(token string) tutorv1connect.TutorServiceClient {
	return tutorv1connect.NewTutorServiceClient(
		connect.NewClient(connecthttp.NewTransport(http.DefaultClient, baseURL), rpcauth.Bearer(token)))
}
