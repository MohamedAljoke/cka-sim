package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/internal/grader"
	"github.com/MohamedAljoke/cka-sim/internal/tasks"
	"github.com/MohamedAljoke/cka-sim/internal/terminal"
)

func TestScoreWeighsPartialCredit(t *testing.T) {
	exam := []tasks.Task{{ID: "a", Weight: 8}, {ID: "b", Weight: 4}, {ID: "c", Weight: 8}}
	results := []grader.Result{
		{TaskID: "a", Earned: 4, Total: 4}, // 8 of 8
		{TaskID: "b", Earned: 1, Total: 2}, // 2 of 4
		// c was never graded: 0 of 8
	}
	if got := Score(exam, results); got != 50 {
		t.Fatalf("score %v, want 50", got)
	}
}

// studyServer serves one fake task, scoring 2 of 3 points, without any cluster.
func studyServer(t *testing.T, study bool) (*httptest.Server, *Store) {
	t.Helper()
	dir := t.TempDir()
	taskDir := filepath.Join(dir, "tasks", "x", "t1")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(dir, "tasks", "lib.sh"): "",
		filepath.Join(taskDir, "check.sh"):    "echo 'PASS 2 it works'; echo 'FAIL 1 not quite'",
		filepath.Join(taskDir, "setup.sh"):    "true",
		filepath.Join(taskDir, "solution.sh"): "echo fixed",
	}
	for path, body := range files {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	store := NewStore(dir)
	if err := store.Save(&Exam{Study: study, StartedAt: time.Now(), TaskIDs: []string{"t1"}}); err != nil {
		t.Fatal(err)
	}
	srv := &Server{
		Store:  store,
		Tasks:  []tasks.Task{{ID: "t1", Weight: 5, Dir: "tasks/x/t1", Explain: "why"}},
		Runner: grader.Runner{AssetsDir: dir, StateDir: dir, Environ: os.Environ()},
		UI:     fstest.MapFS{},
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, store
}

func TestStudyCheckGradesOneTaskAndRevealsSolution(t *testing.T) {
	ts, _ := studyServer(t, true)
	res, err := http.Post(ts.URL+"/api/check/t1", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var v examView
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		t.Fatal(err)
	}
	q := v.Questions[0]
	if q.Result == nil || q.Result.Earned != 2 || q.Result.Total != 3 {
		t.Fatalf("result %+v, want 2/3", q.Result)
	}
	if q.Explain != "why" || q.Solution != "echo fixed" || v.Ended {
		t.Fatalf("study view should reveal help and keep going: %+v", v)
	}
}

func TestExamRefusesStudyActions(t *testing.T) {
	ts, _ := studyServer(t, false)
	for _, path := range []string{"/api/check/t1", "/api/reset/t1"} {
		res, err := http.Post(ts.URL+path, "", nil)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusForbidden {
			t.Fatalf("%s: status %d, want 403", path, res.StatusCode)
		}
	}
	res, err := http.Get(ts.URL + "/api/exam")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var v examView
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		t.Fatal(err)
	}
	if v.Questions[0].Explain != "" || v.Questions[0].Solution != "" {
		t.Fatal("an exam in progress must not reveal explanations or solutions")
	}
}

func TestOtherWebsitesCannotChangeTheExam(t *testing.T) {
	ts, _ := studyServer(t, true)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/end", nil)
	req.Header.Set("Origin", "https://evil.example")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin end: status %d, want 403", res.StatusCode)
	}
	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/api/flag/t1", nil)
	req.Header.Set("Origin", ts.URL)
	if res, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("same-origin flag: status %d, want 200", res.StatusCode)
	}
}

func TestTerminals(t *testing.T) {
	ts, _ := studyServer(t, true)
	if res, err := http.Get(ts.URL + "/api/terminals"); err != nil || res.StatusCode != http.StatusNotFound {
		t.Fatalf("a panel without terminals: %v %v, want 404", res.StatusCode, err)
	}

	srv := &Server{Store: NewStore(t.TempDir()), UI: fstest.MapFS{}, Terminals: &terminal.Manager{
		Command: func(string) *exec.Cmd { return exec.Command("bash", "--norc", "--noprofile", "-i") },
	}}
	t.Cleanup(srv.Terminals.CloseAll)
	ts = httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	res, err := http.Post(ts.URL+"/api/terminals", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	var info terminal.Info
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	// No other website may attach to a shell.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/terminals/" + info.ID + "/ws"
	_, res, err = websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: http.Header{"Origin": {"https://evil.example"}}})
	if err == nil || res == nil || res.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin attach: %v, want 403", err)
	}
	c, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: http.Header{"Origin": {ts.URL}}})
	if err != nil {
		t.Fatalf("same-origin attach: %v", err)
	}
	c.CloseNow()

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/terminals/"+info.ID, nil)
	if res, err = http.DefaultClient.Do(req); err != nil || res.StatusCode != http.StatusNoContent {
		t.Fatalf("close: %v %v, want 204", res.StatusCode, err)
	}
}
