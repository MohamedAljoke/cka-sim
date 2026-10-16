package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

func TestListsTasksWithoutTheQuestion(t *testing.T) {
	status, body := call(t, New([]tasks.Task{wlScale}), http.MethodGet, "/api/tasks")

	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	if !strings.Contains(body, `"id":"wl-scale"`) || !strings.Contains(body, `"title":"Scale a Deployment"`) ||
		!strings.Contains(body, `"domainTitle":"Workloads and Scheduling"`) {
		t.Errorf("body misses the task: %s", body)
	}
	if strings.Contains(body, "Scale it to 4") {
		t.Errorf("body leaks the question: %s", body)
	}
}

func TestListsAnEmptyCatalogAsAnEmptyArray(t *testing.T) {
	_, body := call(t, New(nil), http.MethodGet, "/api/tasks")

	if strings.TrimSpace(body) != "[]" {
		t.Errorf("body = %s, want []", body)
	}
}

func TestTasksCannotBePosted(t *testing.T) {
	status, _ := call(t, New([]tasks.Task{wlScale}), http.MethodPost, "/api/tasks")

	if status != http.StatusMethodNotAllowed {
		t.Errorf("status %d, want 405", status)
	}
}

var wlScale = tasks.Task{ID: "wl-scale", Title: "Scale a Deployment", Domain: tasks.Workloads, Weight: 4, Host: "node-1", Question: "Scale it to 4."}

func call(t *testing.T, h http.Handler, method, path string) (int, string) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	req, err := http.NewRequest(method, srv.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, string(body)
}
