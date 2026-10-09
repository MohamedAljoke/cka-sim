package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/grader"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

func TestListsTasksWithoutAnswers(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := call(t, srv, http.MethodGet, "/api/tasks")

	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	if !strings.Contains(body, `"id":"wl-scale"`) || !strings.Contains(body, `"title":"Scale a Deployment"`) {
		t.Errorf("body misses the task: %s", body)
	}
	if strings.Contains(body, "Scale it to 4") || strings.Contains(body, "kubectl scale") {
		t.Errorf("body leaks the question or explanation: %s", body)
	}
}

func TestQuestionIsTheTaskText(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)

	status, body := call(t, srv, http.MethodGet, "/api/tasks/wl-scale/question")

	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	var reply struct{ Question string }
	if err := json.Unmarshal([]byte(body), &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Question != "Scale it to 4." {
		t.Errorf("question = %q", reply.Question)
	}
	if r.script != "" {
		t.Errorf("reading the question ran %q", r.script)
	}
}

func TestQuestionUnknownTaskIs404(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := call(t, srv, http.MethodGet, "/api/tasks/nope/question")

	if status != http.StatusNotFound || !strings.Contains(body, `no task "nope"`) {
		t.Errorf("got %d %q", status, body)
	}
}

func TestStartRunsSetup(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)

	status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start")

	if status != http.StatusNoContent {
		t.Fatalf("status %d: %s", status, body)
	}
	if r.script != "fresh_ns wl-scale\n" {
		t.Errorf("ran %q, want setup.sh", r.script)
	}
}

func TestStartUnknownTaskIs404(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := call(t, srv, http.MethodPost, "/api/tasks/nope/start")

	if status != http.StatusNotFound || !strings.Contains(body, `no task "nope"`) {
		t.Errorf("got %d %q", status, body)
	}
}

func TestStartWhileSettingUpIs409(t *testing.T) {
	r := &fakeRunner{started: make(chan struct{}), release: make(chan struct{})}
	srv := newAPI(t, r)
	done := make(chan int)
	go func() {
		status, _ := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start")
		done <- status
	}()
	<-r.started

	status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start")

	if status != http.StatusConflict {
		t.Errorf("second start got %d %q, want 409", status, body)
	}
	close(r.release)
	if first := <-done; first != http.StatusNoContent {
		t.Errorf("first start got %d", first)
	}
}

func TestStartShowsSetupFailure(t *testing.T) {
	srv := newAPI(t, &fakeRunner{err: errors.New("namespace stuck")})

	status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start")

	if status != http.StatusInternalServerError || !strings.Contains(body, "namespace stuck") {
		t.Errorf("got %d %q", status, body)
	}
}

func TestCheckReturnsTheScore(t *testing.T) {
	r := &fakeRunner{out: "noise\nPASS 2 scaled\nFAIL 2 ready\n"}
	srv := newAPI(t, r)

	status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/check")

	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	var result grader.Result
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatal(err)
	}
	if result.Earned != 2 || result.Total != 4 || len(result.Checks) != 2 {
		t.Errorf("got %+v, want 2/4 from two checks", result)
	}
	if r.script != "true\n" {
		t.Errorf("ran %q, want check.sh", r.script)
	}
}

func TestCheckUnknownTaskIs404(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := call(t, srv, http.MethodPost, "/api/tasks/nope/check")

	if status != http.StatusNotFound || !strings.Contains(body, `no task "nope"`) {
		t.Errorf("got %d %q", status, body)
	}
}

func TestCheckWhileStartingIs409(t *testing.T) {
	r := &fakeRunner{started: make(chan struct{}), release: make(chan struct{})}
	srv := newAPI(t, r)
	done := make(chan int)
	go func() {
		status, _ := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start")
		done <- status
	}()
	<-r.started

	status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/check")

	if status != http.StatusConflict || !strings.Contains(body, "a script is already running") {
		t.Errorf("check during setup got %d %q, want 409", status, body)
	}
	close(r.release)
	<-done
}

func TestCheckShowsScriptFailure(t *testing.T) {
	srv := newAPI(t, &fakeRunner{err: errors.New("kubectl: not found")})

	status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/check")

	if status != http.StatusInternalServerError || !strings.Contains(body, "kubectl: not found") {
		t.Errorf("got %d %q", status, body)
	}
}

func TestSolutionIsTheExplanation(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)

	status, body := call(t, srv, http.MethodGet, "/api/tasks/wl-scale/solution")

	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	var reply struct{ Explain string }
	if err := json.Unmarshal([]byte(body), &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Explain != "Use kubectl scale." {
		t.Errorf("explain = %q", reply.Explain)
	}
	if r.script != "" {
		t.Errorf("reading the solution ran %q", r.script)
	}
}

func TestRefusesCrossSitePost(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/tasks/wl-scale/start", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden || r.script != "" {
		t.Errorf("got %d and ran %q, want 403 and nothing run", resp.StatusCode, r.script)
	}
}

func newAPI(t *testing.T, r *fakeRunner) *httptest.Server {
	files := fstest.MapFS{
		"wl-scale/task.md":     {Data: []byte("---\nid: wl-scale\ntitle: Scale a Deployment\nhost: node-1\ndomain: workloads\nweight: 4\n---\nScale it to 4.\n")},
		"wl-scale/setup.sh":    {Data: []byte("fresh_ns wl-scale\n")},
		"wl-scale/check.sh":    {Data: []byte("true\n")},
		"wl-scale/solution.sh": {Data: []byte("kubectl scale\n")},
		"wl-scale/explain.md":  {Data: []byte("Use kubectl scale.\n")},
	}
	all, err := tasks.Load(files)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(New(newFakeOpener(), Practice{Tasks: all, Files: files, Runner: r}))
	t.Cleanup(srv.Close)
	return srv
}

func call(t *testing.T, srv *httptest.Server, method, path string) (int, string) {
	req, err := http.NewRequest(method, srv.URL+path, nil)
	if err != nil {
		t.Error(err)
		return 0, ""
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return 0, ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

type fakeRunner struct {
	script, out      string
	err              error
	started, release chan struct{}
}

func (r *fakeRunner) Run(_ context.Context, _, _ string, script []byte) (string, error) {
	r.script = string(script)
	if r.started != nil {
		close(r.started)
		<-r.release
	}
	return r.out, r.err
}
