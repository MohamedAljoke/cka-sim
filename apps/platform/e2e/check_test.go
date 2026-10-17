package e2e

import (
	"net/http"
	"strings"
	"testing"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

func TestCheckingATaskGradesItsCatalogCheckOnItsHost(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, body := request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/check", nil)

	want := `{"earned":2,"total":3,"checks":[{"passed":true,"points":2,"description":"checked on cka-sim-control-plane"},` +
		`{"passed":false,"points":1,"description":"not done yet"}]}`
	if status != http.StatusOK || strings.TrimSpace(string(body)) != want {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestCheckingATaskBeforeTheLabIsReadyIsAConflict(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})

	status, body := request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/check", nil)

	if status != http.StatusConflict || !strings.Contains(string(body), `"code":"lab_not_ready"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestCheckingAnUnknownTaskIsNotFound(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, body := request(t, http.MethodPost, "/api/tasks/no-such-task/check", nil)

	if status != http.StatusNotFound || !strings.Contains(string(body), `"code":"task_not_found"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestACrossSiteTaskCheckIsRefused(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, _ := request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/check", map[string]string{"Sec-Fetch-Site": "cross-site"})

	if status != http.StatusForbidden {
		t.Errorf("status %d, want 403", status)
	}
}
