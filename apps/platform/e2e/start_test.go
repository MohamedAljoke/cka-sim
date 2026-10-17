package e2e

import (
	"net/http"
	"strings"
	"testing"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

func TestStartingATaskRunsItsCatalogSetupOnItsHost(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, body := request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/start", nil)

	if status != http.StatusOK || strings.TrimSpace(string(body)) != `{"passed":true,"exitCode":0,"output":"ran on cka-sim-control-plane\n"}` {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestStartingATaskBeforeTheLabIsReadyIsAConflict(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})

	status, body := request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/start", nil)

	if status != http.StatusConflict || !strings.Contains(string(body), `"code":"lab_not_ready"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestStartingAnUnknownTaskIsNotFound(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, body := request(t, http.MethodPost, "/api/tasks/no-such-task/start", nil)

	if status != http.StatusNotFound || !strings.Contains(string(body), `"code":"task_not_found"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestACrossSiteTaskStartIsRefused(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, _ := request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/start", map[string]string{"Sec-Fetch-Site": "cross-site"})

	if status != http.StatusForbidden {
		t.Errorf("status %d, want 403", status)
	}
}
