package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

func TestNoExamIsNotFound(t *testing.T) {
	url := ownServer(t, t.TempDir())

	status, body := requestTo(t, url, http.MethodGet, "/api/exam", nil)

	if status != http.StatusNotFound || !strings.Contains(string(body), `"code":"no_exam"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestABadExamRequestIsRefused(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	url := ownServer(t, t.TempDir())

	for _, tc := range []struct{ body, code string }{
		{`{"count": 0, "minutes": 30}`, "invalid_exam"},
		{`{"count": 2`, "malformed_request"},
	} {
		status, body := post(t, url, "/api/exam", tc.body)

		if status != http.StatusBadRequest || !strings.Contains(body, `"code":"`+tc.code+`"`) {
			t.Errorf("%s: status %d: %s", tc.body, status, body)
		}
	}
}

func TestAnExamNeedsAReadyLab(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})
	url := ownServer(t, t.TempDir())

	status, body := post(t, url, "/api/exam", `{"count": 2, "minutes": 30}`)

	if status != http.StatusConflict || !strings.Contains(body, `"code":"lab_not_ready"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestAnExamIsSetUpPlayedGradedAndDiscarded(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	url := ownServer(t, t.TempDir())

	status, body := post(t, url, "/api/exam", `{"count": 2, "minutes": 30}`)
	if status != http.StatusAccepted {
		t.Fatalf("begin: status %d: %s", status, body)
	}
	e := waitForExam(t, url, func(e exam) bool { return e.Deadline != "" })
	if len(e.Tasks) != 2 || e.Tasks[0].Setup != "ready" || e.Tasks[1].Setup != "ready" {
		t.Fatalf("set up: %+v", e)
	}

	if status, body := requestTo(t, url, http.MethodPost, "/api/tasks/"+e.Tasks[0].ID+"/check", nil); status != http.StatusConflict ||
		!strings.Contains(string(body), `"code":"exam_running"`) {
		t.Errorf("practice check during the exam: status %d: %s", status, body)
	}
	if status, body := requestTo(t, url, http.MethodPut, "/api/exam/tasks/"+e.Tasks[1].ID+"/flag", nil); status != http.StatusNoContent {
		t.Errorf("flag: status %d: %s", status, body)
	}

	if status, body := requestTo(t, url, http.MethodPost, "/api/exam/end", nil); status != http.StatusAccepted {
		t.Fatalf("end: status %d: %s", status, body)
	}
	e = waitForExam(t, url, func(e exam) bool { return e.Score != nil })
	if e.Score.Percent != 67 || !e.Score.Passed || !e.Tasks[1].Flagged || e.Tasks[0].Grade == nil || e.Tasks[0].Grade.Earned != 2 {
		t.Errorf("graded: %+v", e)
	}

	if status, body := requestTo(t, url, http.MethodDelete, "/api/exam", nil); status != http.StatusNoContent {
		t.Errorf("discard: status %d: %s", status, body)
	}
	if status, _ := requestTo(t, url, http.MethodGet, "/api/exam", nil); status != http.StatusNotFound {
		t.Errorf("after discard: status %d", status)
	}
}

func TestAnExamOutlivesAServerRestart(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	dir := t.TempDir()
	server, url, err := start(agentURL, dir)
	if err != nil {
		t.Fatal(err)
	}
	post(t, url, "/api/exam", `{"count": 2, "minutes": 30}`)
	before := waitForExam(t, url, func(e exam) bool { return e.Deadline != "" })
	if err := stop(server); err != nil {
		t.Fatal(err)
	}

	url = ownServer(t, dir)

	after := waitForExam(t, url, func(e exam) bool { return true })
	if after.Deadline != before.Deadline || len(after.Tasks) != 2 || after.Tasks[0].ID != before.Tasks[0].ID {
		t.Errorf("before %+v\nafter %+v", before, after)
	}
}

func ownServer(t *testing.T, stateDir string) string {
	t.Helper()
	server, url, err := start(agentURL, stateDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := stop(server); err != nil {
			t.Error(err)
		}
	})
	return url
}

func post(t *testing.T, url, path, body string) (int, string) {
	t.Helper()
	res, err := http.Post(url+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, string(got)
}

func waitForExam(t *testing.T, url string, done func(exam) bool) exam {
	t.Helper()
	for range 100 {
		status, body := requestTo(t, url, http.MethodGet, "/api/exam", nil)
		var e exam
		if status == http.StatusOK && json.Unmarshal(body, &e) == nil && done(e) {
			return e
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("the exam never got there")
	return exam{}
}

type exam struct {
	Tasks []struct {
		ID      string `json:"id"`
		Setup   string `json:"setup"`
		Flagged bool   `json:"flagged"`
		Grade   *struct {
			Earned int `json:"earned"`
			Total  int `json:"total"`
		} `json:"grade"`
	} `json:"tasks"`
	Deadline string `json:"deadline"`
	Score    *struct {
		Percent int  `json:"percent"`
		Passed  bool `json:"passed"`
	} `json:"score"`
}
