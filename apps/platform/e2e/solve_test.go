package e2e

import (
	"net/http"
	"strings"
	"testing"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

func TestSolvingATaskRunsItsCatalogSolutionOnItsHost(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	ran := agent.scriptsFrom()

	status, body := request(t, http.MethodPost, "/api/tasks/nw-nodeport/solve", nil)

	scripts := ran()
	if status != http.StatusOK || strings.TrimSpace(string(body)) != `{"passed":true,"exitCode":0,"output":"ran on cka-sim-control-plane\n"}` ||
		len(scripts) != 1 || !strings.Contains(scripts[0], "create service nodeport web --tcp=80:80 --node-port=30080") {
		t.Errorf("status %d: %s; ran %q", status, body, scripts)
	}
}

func TestSolvingATaskBeforeTheLabIsReadyIsAConflict(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})

	status, body := request(t, http.MethodPost, "/api/tasks/nw-nodeport/solve", nil)

	if status != http.StatusConflict || !strings.Contains(string(body), `"code":"lab_not_ready"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestSolvingAnUnknownTaskIsNotFound(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, body := request(t, http.MethodPost, "/api/tasks/no-such-task/solve", nil)

	if status != http.StatusNotFound || !strings.Contains(string(body), `"code":"task_not_found"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestACrossSiteSolveIsRefused(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, _ := request(t, http.MethodPost, "/api/tasks/nw-nodeport/solve", map[string]string{"Sec-Fetch-Site": "cross-site"})

	if status != http.StatusForbidden {
		t.Errorf("status %d, want 403", status)
	}
}
