package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/exam"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/grader"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

func TestListsTasksWithoutAnswers(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := call(t, srv, http.MethodGet, "/api/tasks")

	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	if !strings.Contains(body, `"id":"wl-scale"`) || !strings.Contains(body, `"title":"Scale a Deployment"`) ||
		!strings.Contains(body, `"domainTitle":"Workloads and Scheduling"`) {
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
	if ran := r.ran(); len(ran) != 0 {
		t.Errorf("reading the question ran %q", ran)
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
	if ran, want := r.ran(), []string{"heal node-1\n", "fresh_ns wl-scale\n"}; !slices.Equal(ran, want) {
		t.Errorf("ran %q, want reset.sh then setup.sh", ran)
	}
}

func TestLeavingPracticeTidiesTheTask(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)
	r.pause()

	status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/tidy")

	if status != http.StatusAccepted {
		t.Fatalf("status %d: %s", status, body)
	}
	r.resume()
	if status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start"); status != http.StatusNoContent {
		t.Fatalf("the next start got %d %q, want it to wait for the tidy", status, body)
	}
	if ran := r.ran(); len(ran) < 2 || ran[0] != "heal node-1\n" || ran[1] != tasks.TidyScript {
		t.Errorf("ran %q, want reset.sh then the tidy first", ran)
	}
}

func TestTidyWithoutALabDoesNothing(t *testing.T) {
	r := &fakeRunner{}
	srv, _ := newAPIWithoutLab(t, r)

	status, _ := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/tidy")

	if status != http.StatusNoContent || len(r.ran()) != 0 {
		t.Errorf("got %d and ran %q, want 204 and nothing run", status, r.ran())
	}
}

func TestCancellingStartStopsTheScript(t *testing.T) {
	r := &fakeRunner{started: make(chan struct{}), release: make(chan struct{})}
	srv := newAPI(t, r)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL+"/api/tasks/wl-scale/start", nil)
		if resp, err := http.DefaultClient.Do(req); err == nil {
			resp.Body.Close()
		}
	}()
	<-r.started

	cancel()
	<-done

	status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start")
	if status != http.StatusNoContent {
		t.Errorf("the next start got %d %q, want the lock freed by the cancelled one", status, body)
	}
	if !r.wasCancelled() {
		t.Error("the runner's context was not cancelled")
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
	if ran, want := r.ran(), []string{"true\n"}; !slices.Equal(ran, want) {
		t.Errorf("ran %q, want check.sh", ran)
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
	if ran := r.ran(); len(ran) != 0 {
		t.Errorf("reading the solution ran %q", ran)
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

	if ran := r.ran(); resp.StatusCode != http.StatusForbidden || len(ran) != 0 {
		t.Errorf("got %d and ran %q, want 403 and nothing run", resp.StatusCode, ran)
	}
}

func newAPI(t *testing.T, r *fakeRunner) *httptest.Server {
	srv, lab := newAPIWithoutLab(t, r)
	if err := lab.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
	return srv
}

func newAPIWithoutLab(t *testing.T, r *fakeRunner) (*httptest.Server, *sandbox.Lab) {
	lab := sandbox.NewLab(fakeProvider{runner: r, shells: newFakeOpener()}, time.Now)
	lockWait = 50 * time.Millisecond
	files := fstest.MapFS{
		"wl-scale/task.md":     {Data: []byte("---\nid: wl-scale\ntitle: Scale a Deployment\nhost: node-1\ndomain: workloads\ntopics: [deployments]\nweight: 4\n---\nScale it to 4.\n")},
		"wl-scale/setup.sh":    {Data: []byte("fresh_ns wl-scale\n")},
		"wl-scale/check.sh":    {Data: []byte("true\n")},
		"wl-scale/solution.sh": {Data: []byte("kubectl scale\n")},
		"wl-scale/explain.md":  {Data: []byte("Use kubectl scale.\n")},
		"wl-scale/reset.sh":    {Data: []byte("heal node-1\n")},
	}
	all, err := tasks.Load(files)
	if err != nil {
		t.Fatal(err)
	}
	session, err := exam.Open(exam.Config{
		Store:   exam.FileStore{Path: filepath.Join(t.TempDir(), "exam.json")},
		Files:   files,
		Runner:  lab,
		Catalog: all,
		Now:     time.Now,
		Rand:    rand.New(rand.NewPCG(1, 2)),
		Ready:   lab.Ensure,
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(New(lab, Practice{Tasks: all, Files: files, Exam: session}))
	t.Cleanup(srv.Close)
	return srv, lab
}

func readyLab(t *testing.T, r *fakeRunner, shells *fakeOpener) *sandbox.Lab {
	lab := sandbox.NewLab(fakeProvider{runner: r, shells: shells}, time.Now)
	if err := lab.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
	return lab
}

func call(t *testing.T, srv *httptest.Server, method, path string) (int, string) {
	return request(t, srv, method, path, "")
}

func request(t *testing.T, srv *httptest.Server, method, path, payload string) (int, string) {
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(payload))
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

// fakeRunner records every script body it runs. With started set, the first run signals started
// and waits for release or a cancel. While paused, every run waits for resume.
type fakeRunner struct {
	out              string
	err              error
	started, release chan struct{}
	mu               sync.Mutex
	gate             chan struct{}
	// only, when set, makes pause hold just this script.
	only      string
	scripts   []string
	cancelled bool
}

func (r *fakeRunner) Run(ctx context.Context, _, _ string, script []byte) (string, error) {
	r.mu.Lock()
	r.scripts = append(r.scripts, string(script))
	first := len(r.scripts) == 1
	gate := r.gate
	r.mu.Unlock()
	if gate != nil && (r.only == "" || r.only == string(script)) {
		<-gate
	}
	if r.started != nil && first {
		close(r.started)
		select {
		case <-r.release:
		case <-ctx.Done():
			r.mu.Lock()
			r.cancelled = true
			r.mu.Unlock()
			return "", ctx.Err()
		}
	}
	return r.out, r.err
}

func (r *fakeRunner) pause() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gate = make(chan struct{})
}

func (r *fakeRunner) resume() { close(r.gate) }

func (r *fakeRunner) wasCancelled() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cancelled
}

func (r *fakeRunner) ran() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.scripts)
}

type fakeProvider struct {
	runner *fakeRunner
	shells *fakeOpener
}

func (fakeProvider) Name() string { return "Fake" }

func (p fakeProvider) Create(context.Context) (sandbox.Box, error) {
	return sandbox.Box{ID: "fake", Runner: p.runner, Shells: p.shells}, nil
}

func (fakeProvider) Destroy(context.Context, sandbox.Box) error { return nil }
