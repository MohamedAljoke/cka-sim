package e2e

import (
	"net/http"
	"strings"
	"testing"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

func TestTheCurrentTaskIsTheOneSetUpLast(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/start", nil)

	status, body := request(t, http.MethodGet, "/api/tasks/current", nil)

	if status != http.StatusOK || !strings.Contains(string(body), `"id":"ar-aggregate-view"`) ||
		!strings.Contains(string(body), `"setup":"ready"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestATidiedTaskIsNoLongerCurrent(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/start", nil)
	request(t, http.MethodPost, "/api/tasks/ar-aggregate-view/tidy", nil)

	status, body := request(t, http.MethodGet, "/api/tasks/current", nil)

	if status != http.StatusNotFound || !strings.Contains(string(body), `"code":"no_current_task"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestTheCurrentTaskIsRefusedWhileAnExamRuns(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	url := ownServer(t, t.TempDir())
	if status, body := post(t, url, "/api/exam", `{"count": 1, "minutes": 30}`); status != http.StatusAccepted {
		t.Fatalf("begin: status %d: %s", status, body)
	}

	status, body := requestTo(t, url, http.MethodGet, "/api/tasks/current", nil)

	if status != http.StatusConflict || !strings.Contains(string(body), `"code":"exam_running"`) {
		t.Errorf("status %d: %s", status, body)
	}
}
