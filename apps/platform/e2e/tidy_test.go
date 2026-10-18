package e2e

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

func TestTidyingATaskResetsItThenDeletesItsNamespaces(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	ran := agent.scriptsFrom()

	status, body := request(t, http.MethodPost, "/api/tasks/wl-taints/tidy", nil)

	scripts := ran()
	if status != http.StatusNoContent || len(scripts) != 2 || !strings.Contains(scripts[0], "dedicated-") ||
		!slices.Contains(scripts, "kubectl delete namespace -l cka-sim/task=wl-taints --wait=false >/dev/null\n") {
		t.Errorf("status %d: %s; ran %q", status, body, scripts)
	}
}

func TestTidyingATaskBeforeTheLabIsReadyIsAConflict(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})

	status, body := request(t, http.MethodPost, "/api/tasks/wl-taints/tidy", nil)

	if status != http.StatusConflict || !strings.Contains(string(body), `"code":"lab_not_ready"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestTidyingAnUnknownTaskIsNotFound(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, body := request(t, http.MethodPost, "/api/tasks/no-such-task/tidy", nil)

	if status != http.StatusNotFound || !strings.Contains(string(body), `"code":"task_not_found"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestACrossSiteTidyIsRefused(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	status, _ := request(t, http.MethodPost, "/api/tasks/wl-taints/tidy", map[string]string{"Sec-Fetch-Site": "cross-site"})

	if status != http.StatusForbidden {
		t.Errorf("status %d, want 403", status)
	}
}
