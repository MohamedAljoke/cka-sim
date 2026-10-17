package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

func TestListsTasksWithoutTheQuestion(t *testing.T) {
	status, body := call(t, newAPI(wlScale), http.MethodGet, "/api/tasks")

	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	if !strings.Contains(body, `"id":"wl-scale"`) || !strings.Contains(body, `"title":"Scale a Deployment"`) ||
		!strings.Contains(body, `"domainTitle":"Workloads and Scheduling"`) {
		t.Errorf("body misses the task: %s", body)
	}
	if strings.Contains(body, "Scale it to 4") || strings.Contains(body, "kubectl scale") {
		t.Errorf("body leaks the question or solution: %s", body)
	}
}

func TestATaskWithoutTopicsListsAnEmptyArray(t *testing.T) {
	task := wlScale
	task.Topics = nil

	_, body := call(t, newAPI(task), http.MethodGet, "/api/tasks")

	if !strings.Contains(body, `"topics":[]`) {
		t.Errorf("body = %s, want topics as []", body)
	}
}

func TestListsAnEmptyCatalogAsAnEmptyArray(t *testing.T) {
	_, body := call(t, newAPI(), http.MethodGet, "/api/tasks")

	if strings.TrimSpace(body) != "[]" {
		t.Errorf("body = %s, want []", body)
	}
}

func TestFiltersTheListByDomainAndTopic(t *testing.T) {
	netpol := tasks.Task{ID: "nw-netpol", Title: "Lock down", Domain: tasks.Networking, Topics: []string{"networkpolicy"}, Weight: 4, Host: "node-1"}
	dns := tasks.Task{ID: "nw-dns", Title: "Fix DNS", Domain: tasks.Networking, Topics: []string{"dns"}, Weight: 4, Host: "node-1"}

	status, body := call(t, newAPI(wlScale, netpol, dns), http.MethodGet, "/api/tasks?domain=networking&topic=networkpolicy&topic=egress")

	if status != http.StatusOK || !strings.Contains(body, `"id":"nw-netpol"`) ||
		strings.Contains(body, `"id":"nw-dns"`) || strings.Contains(body, `"id":"wl-scale"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestAFilterMatchingNothingListsAnEmptyArray(t *testing.T) {
	status, body := call(t, newAPI(wlScale), http.MethodGet, "/api/tasks?domain=storage")

	if status != http.StatusOK || strings.TrimSpace(body) != "[]" {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestABadFilterIsABadRequest(t *testing.T) {
	for _, tc := range []struct{ query, code string }{
		{"domain=security", "unknown_domain"},
		{"domain=", "unknown_domain"},
		{"topic=Network%20Policy", "malformed_topic"},
	} {
		status, body := call(t, newAPI(wlScale), http.MethodGet, "/api/tasks?"+tc.query)

		if status != http.StatusBadRequest || !strings.Contains(body, `"code":"`+tc.code+`"`) {
			t.Errorf("%s: status %d: %s", tc.query, status, body)
		}
	}
}

func TestTasksCannotBePosted(t *testing.T) {
	status, _ := call(t, newAPI(wlScale), http.MethodPost, "/api/tasks")

	if status != http.StatusMethodNotAllowed {
		t.Errorf("status %d, want 405", status)
	}
}

func TestAFailedListIsUnavailableAndHidesTheCause(t *testing.T) {
	h := newAPIWith(fakeRepository{err: errors.New("disk gone at /secret/path")})

	status, body := call(t, h, http.MethodGet, "/api/tasks")

	if status != http.StatusServiceUnavailable || !strings.Contains(body, `"code":"tasks_unavailable"`) {
		t.Errorf("status %d: %s", status, body)
	}
	if strings.Contains(body, "secret") {
		t.Errorf("body leaks the cause: %s", body)
	}
}

func TestATaskComesWithItsQuestion(t *testing.T) {
	status, body := call(t, newAPI(wlScale), http.MethodGet, "/api/tasks/wl-scale")

	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	if !strings.Contains(body, `"id":"wl-scale"`) || !strings.Contains(body, `"domainTitle":"Workloads and Scheduling"`) ||
		!strings.Contains(body, `"question":"Scale it to 4."`) {
		t.Errorf("body = %s", body)
	}
}

func TestATaskDoesNotGiveAwayItsSolution(t *testing.T) {
	_, body := call(t, newAPI(wlScale), http.MethodGet, "/api/tasks/wl-scale")

	if strings.Contains(body, "kubectl scale") {
		t.Errorf("body leaks the solution: %s", body)
	}
}

func TestASolutionComesOnlyWhenAskedFor(t *testing.T) {
	status, body := call(t, newAPI(wlScale), http.MethodGet, "/api/tasks/wl-scale/solution")

	if status != http.StatusOK || strings.TrimSpace(body) != `{"solution":"Run kubectl scale."}` {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestTheSolutionOfAnUnknownTaskIsNotFound(t *testing.T) {
	status, body := call(t, newAPI(wlScale), http.MethodGet, "/api/tasks/net-svc/solution")

	if status != http.StatusNotFound || !strings.Contains(body, `"code":"task_not_found"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestAnUnknownTaskIsNotFound(t *testing.T) {
	status, body := call(t, newAPI(wlScale), http.MethodGet, "/api/tasks/net-svc")

	if status != http.StatusNotFound || !strings.Contains(body, `"code":"task_not_found"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestAFailedGetIsUnavailableAndHidesTheCause(t *testing.T) {
	h := newAPIWith(fakeRepository{err: errors.New("disk gone at /secret/path")})

	status, body := call(t, h, http.MethodGet, "/api/tasks/wl-scale")

	if status != http.StatusServiceUnavailable || strings.Contains(body, "secret") {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestStartingATaskRunsItsSetup(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "namespace/wl-scale created\n"}}

	status, body := call(t, newAPIRunning(r, wlScale), http.MethodPost, "/api/tasks/wl-scale/start")

	if status != http.StatusOK || strings.TrimSpace(body) != `{"passed":true,"exitCode":0,"output":"namespace/wl-scale created\n"}` {
		t.Errorf("status %d: %s", status, body)
	}
	if r.host != "node-1" || r.script != wlScale.Setup {
		t.Errorf("ran %q on %q", r.script, r.host)
	}
}

func TestAFailedSetupAnswersItsExitCodeAndOutput(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "error: the server doesn't have a resource type \"widgets\"\n", ExitCode: 1}}

	status, body := call(t, newAPIRunning(r, wlScale), http.MethodPost, "/api/tasks/wl-scale/start")

	if status != http.StatusOK || !strings.Contains(body, `"passed":false,"exitCode":1`) || !strings.Contains(body, "widgets") {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestStartingATaskWhileTheLabIsNotReadyIsAConflict(t *testing.T) {
	r := &fakeRunner{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}

	status, body := call(t, newAPIRunning(r, wlScale), http.MethodPost, "/api/tasks/wl-scale/start")

	if status != http.StatusConflict || !strings.Contains(body, `"code":"lab_not_ready"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestStartingATaskOnALabThatDoesNotAnswerIsUnavailableAndHidesTheCause(t *testing.T) {
	r := &fakeRunner{err: errors.New("dial tcp 127.0.0.1:7171: connection refused")}

	status, body := call(t, newAPIRunning(r, wlScale), http.MethodPost, "/api/tasks/wl-scale/start")

	if status != http.StatusServiceUnavailable || !strings.Contains(body, `"code":"lab_unavailable"`) || strings.Contains(body, "7171") {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestStartingAnUnknownTaskIsNotFound(t *testing.T) {
	r := &fakeRunner{}

	status, body := call(t, newAPIRunning(r, wlScale), http.MethodPost, "/api/tasks/net-svc/start")

	if status != http.StatusNotFound || !strings.Contains(body, `"code":"task_not_found"`) || r.host != "" {
		t.Errorf("status %d: %s", status, body)
	}
}

var wlScale = tasks.Task{ID: "wl-scale", Title: "Scale a Deployment", Domain: tasks.Workloads, Topics: []string{"deployments"}, Weight: 4, Host: "node-1", Question: "Scale it to 4.", Solution: "Run kubectl scale.", Setup: "fresh_ns wl-scale\n", Check: "check 1 \"scaled\" true\n"}

func newAPI(all ...tasks.Task) http.Handler {
	return newAPIWith(fakeRepository{tasks: all})
}

func newAPIWith(repo usecases.TaskRepository) http.Handler {
	return New(usecases.New(repo, &fakeRunner{}), labusecases.UseCases{})
}

func newAPIRunning(r *fakeRunner, all ...tasks.Task) http.Handler {
	return New(usecases.New(fakeRepository{tasks: all}, r), labusecases.UseCases{})
}

type fakeRunner struct {
	result labs.ScriptResult
	err    error
	host   string
	script string
}

func (r *fakeRunner) RunScript(_ context.Context, host, script string) (labs.ScriptResult, error) {
	r.host, r.script = host, script
	return r.result, r.err
}

type fakeRepository struct {
	tasks []tasks.Task
	err   error
}

func (r fakeRepository) List(context.Context) ([]tasks.Task, error) { return r.tasks, r.err }

func (r fakeRepository) Get(_ context.Context, id string) (tasks.Task, error) {
	if r.err != nil {
		return tasks.Task{}, r.err
	}
	i := slices.IndexFunc(r.tasks, func(t tasks.Task) bool { return t.ID == id })
	if i < 0 {
		return tasks.Task{}, tasks.ErrNotFound
	}
	return r.tasks[i], nil
}

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
