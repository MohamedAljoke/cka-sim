package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

func TestListsTasksWithoutTheQuestion(t *testing.T) {
	status, body := call(t, newAPI(t, wlScale), http.MethodGet, "/api/tasks")

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

func TestATaskWithoutTopicsListsAnEmptyArray(t *testing.T) {
	task := wlScale
	task.Topics = nil

	_, body := call(t, newAPI(t, task), http.MethodGet, "/api/tasks")

	if !strings.Contains(body, `"topics":[]`) {
		t.Errorf("body = %s, want topics as []", body)
	}
}

func TestListsAnEmptyCatalogAsAnEmptyArray(t *testing.T) {
	_, body := call(t, newAPI(t), http.MethodGet, "/api/tasks")

	if strings.TrimSpace(body) != "[]" {
		t.Errorf("body = %s, want []", body)
	}
}

func TestTasksCannotBePosted(t *testing.T) {
	status, _ := call(t, newAPI(t, wlScale), http.MethodPost, "/api/tasks")

	if status != http.StatusMethodNotAllowed {
		t.Errorf("status %d, want 405", status)
	}
}

func TestAFailedListIsUnavailableAndHidesTheCause(t *testing.T) {
	h := New(usecases.NewListTasks(fakeRepository{err: errors.New("disk gone at /secret/path")}))

	status, body := call(t, h, http.MethodGet, "/api/tasks")

	if status != http.StatusServiceUnavailable || !strings.Contains(body, `"code":"tasks_unavailable"`) {
		t.Errorf("status %d: %s", status, body)
	}
	if strings.Contains(body, "secret") {
		t.Errorf("body leaks the cause: %s", body)
	}
}

var wlScale = tasks.Task{ID: "wl-scale", Title: "Scale a Deployment", Domain: tasks.Workloads, Topics: []string{"deployments"}, Weight: 4, Host: "node-1", Question: "Scale it to 4."}

func newAPI(t *testing.T, all ...tasks.Task) http.Handler {
	t.Helper()
	return New(usecases.NewListTasks(fakeRepository{tasks: all}))
}

type fakeRepository struct {
	tasks []tasks.Task
	err   error
}

func (r fakeRepository) List(context.Context) ([]tasks.Task, error) { return r.tasks, r.err }

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
