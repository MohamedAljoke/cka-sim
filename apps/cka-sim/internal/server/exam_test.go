package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

func TestNoExamIs404(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := call(t, srv, http.MethodGet, "/api/exam")

	if status != http.StatusNotFound || !strings.Contains(body, "no exam") {
		t.Errorf("got %d %q", status, body)
	}
}

func TestBeginExamSetsUpThenStartsTheClock(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := request(t, srv, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)

	if status != http.StatusAccepted {
		t.Fatalf("status %d: %s", status, body)
	}
	state := waitPrepared(t, srv)
	if len(state.Exam.Tasks) != 1 || state.Exam.Tasks[0].ID != "wl-scale" || state.Exam.Tasks[0].Setup != "ready" {
		t.Errorf("tasks %+v, want wl-scale ready", state.Exam.Tasks)
	}
	if got := state.Exam.Deadline.Sub(state.Exam.Started); got != 30*time.Minute {
		t.Errorf("deadline is %v after the start, want 30m", got)
	}
	if state.Now.IsZero() || state.Score != nil {
		t.Errorf("now %v score %+v, want the server's time and no score yet", state.Now, state.Score)
	}
}

func TestBeginExamRejects(t *testing.T) {
	tests := []struct {
		name, body string
	}{
		{"not JSON", "sixteen"},
		{"no tasks", `{"count": 0, "minutes": 30}`},
		{"no time", `{"count": 1, "minutes": 0}`},
		{"too long", `{"count": 1, "minutes": 600}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newAPI(t, &fakeRunner{})

			status, body := request(t, srv, http.MethodPost, "/api/exam", tt.body)

			if status != http.StatusBadRequest {
				t.Errorf("got %d %q, want 400", status, body)
			}
		})
	}
}

func TestBeginExamTwiceIs409(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	beginExam(t, srv)

	status, body := request(t, srv, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)

	if status != http.StatusConflict || !strings.Contains(body, "already in progress") {
		t.Errorf("got %d %q", status, body)
	}
}

func TestBeginExamWhilePracticeSetsUpIs409(t *testing.T) {
	r := &fakeRunner{started: make(chan struct{}), release: make(chan struct{})}
	srv := newAPI(t, r)
	done := make(chan int)
	go func() {
		status, _ := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start")
		done <- status
	}()
	<-r.started

	status, body := request(t, srv, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)

	if status != http.StatusConflict || !strings.Contains(body, "a script is already running") {
		t.Errorf("got %d %q", status, body)
	}
	close(r.release)
	<-done
}

func TestPracticeIsRefusedDuringAnExam(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	beginExam(t, srv)

	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/api/tasks/wl-scale/start"},
		{http.MethodPost, "/api/tasks/wl-scale/check"},
		{http.MethodGet, "/api/tasks/wl-scale/solution"},
		{http.MethodPost, "/api/tasks/wl-scale/tidy"},
	} {
		status, body := call(t, srv, route.method, route.path)

		if status != http.StatusConflict || !strings.Contains(body, "not during an exam") {
			t.Errorf("%s %s got %d %q", route.method, route.path, status, body)
		}
	}
	if status, _ := call(t, srv, http.MethodGet, "/api/tasks/wl-scale/question"); status != http.StatusOK {
		t.Errorf("the question got %d, want it readable during the exam", status)
	}
}

func TestFlagAQuestion(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	beginExam(t, srv)

	if status, body := call(t, srv, http.MethodPut, "/api/exam/flags/wl-scale"); status != http.StatusNoContent {
		t.Fatalf("flag got %d %q", status, body)
	}
	if !getState(t, srv).Exam.Tasks[0].Flagged {
		t.Error("not flagged after PUT")
	}
	call(t, srv, http.MethodDelete, "/api/exam/flags/wl-scale")
	if getState(t, srv).Exam.Tasks[0].Flagged {
		t.Error("still flagged after DELETE")
	}
	if status, _ := call(t, srv, http.MethodPut, "/api/exam/flags/nope"); status != http.StatusConflict {
		t.Errorf("flagging a task outside the exam got %d", status)
	}
}

func TestEndExamScoresAndUnlocksTheSolution(t *testing.T) {
	srv := newAPI(t, &fakeRunner{out: "PASS 2 scaled\nFAIL 2 ready\n"})
	beginExam(t, srv)

	status, body := call(t, srv, http.MethodPost, "/api/exam/end")

	if status != http.StatusAccepted {
		t.Fatalf("status %d: %s", status, body)
	}
	state := waitScored(t, srv)
	if r := state.Exam.Tasks[0].Result; r == nil || r.Earned != 2 || r.Total != 4 {
		t.Errorf("result %+v, want 2/4", r)
	}
	if state.Score == nil || state.Score.Percent != 50 || state.Score.Passed {
		t.Errorf("score %+v, want 50%% and a fail", state.Score)
	}
	if status, _ := call(t, srv, http.MethodGet, "/api/tasks/wl-scale/solution"); status != http.StatusOK {
		t.Errorf("the solution got %d after the exam, want 200", status)
	}
}

func TestEndExamEndsTheLab(t *testing.T) {
	srv := newAPI(t, &fakeRunner{out: "PASS 4 scaled\n"})
	beginExam(t, srv)
	waitPrepared(t, srv)

	if status, body := call(t, srv, http.MethodPost, "/api/exam/end"); status != http.StatusAccepted {
		t.Fatalf("status %d: %s", status, body)
	}

	waitScored(t, srv)
	waitLab(t, srv, sandbox.None)
}

func TestDiscardExam(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	beginExam(t, srv)

	if status, body := call(t, srv, http.MethodDelete, "/api/exam"); status != http.StatusNoContent {
		t.Fatalf("got %d %q", status, body)
	}

	if status, _ := call(t, srv, http.MethodGet, "/api/exam"); status != http.StatusNotFound {
		t.Errorf("GET after discard got %d, want 404", status)
	}
	if got := getLab(t, srv).State; got != sandbox.None {
		t.Errorf("lab is %s after discarding, want %s", got, sandbox.None)
	}
}

func TestBeginAFocusedExam(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	for _, payload := range []string{
		`{"count": 1, "minutes": 30, "domains": ["storage"]}`,
		`{"count": 1, "minutes": 30, "topics": ["rbac"]}`,
	} {
		if status, body := request(t, srv, http.MethodPost, "/api/exam", payload); status != http.StatusBadRequest {
			t.Errorf("%s got %d %q, want 400", payload, status, body)
		}
	}
	payload := `{"count": 1, "minutes": 30, "domains": ["workloads"], "topics": ["deployments"]}`
	if status, body := request(t, srv, http.MethodPost, "/api/exam", payload); status != http.StatusAccepted {
		t.Fatalf("got %d %q", status, body)
	}
	if state := waitPrepared(t, srv); state.Exam.Tasks[0].ID != "wl-scale" {
		t.Errorf("tasks = %+v, want wl-scale", state.Exam.Tasks)
	}
}

func TestScoreWaitsForEveryCheck(t *testing.T) {
	r := &fakeRunner{out: "PASS 4 scaled\n"}
	srv := newAPI(t, r)
	beginExam(t, srv)
	r.pause()

	status, body := call(t, srv, http.MethodPost, "/api/exam/end")

	if status != http.StatusAccepted {
		t.Fatalf("status %d: %s", status, body)
	}
	if state := getState(t, srv); state.Score != nil || !state.Exam.Scoring() {
		t.Errorf("got %+v, want scoring and no score yet", state)
	}
	r.resume()
	if state := waitScored(t, srv); state.Score == nil || state.Score.Percent != 100 {
		t.Errorf("score %+v, want 100%%", state.Score)
	}
}

func TestScoreDoesNotWaitForTidy(t *testing.T) {
	r := &fakeRunner{out: "PASS 4 scaled\n", only: tasks.TidyScript}
	srv := newAPI(t, r)
	beginExam(t, srv)
	r.pause()
	defer r.resume()

	call(t, srv, http.MethodPost, "/api/exam/end")

	if state := waitScored(t, srv); state.Score.Percent != 100 || !state.Exam.Scoring() {
		t.Errorf("got %+v, want 100%% while the tidy still runs", state)
	}
}

func beginExam(t *testing.T, srv *httptest.Server) {
	if status, body := request(t, srv, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`); status != http.StatusAccepted {
		t.Fatalf("begin got %d %q", status, body)
	}
	waitPrepared(t, srv)
}

func waitPrepared(t *testing.T, srv *httptest.Server) examState {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if state := getState(t, srv); state.Exam.Prepared() {
			return state
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the exam never finished setting up")
	return examState{}
}

func waitScored(t *testing.T, srv *httptest.Server) examState {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if state := getState(t, srv); state.Score != nil {
			return state
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the exam was never scored")
	return examState{}
}

func getState(t *testing.T, srv *httptest.Server) examState {
	status, body := call(t, srv, http.MethodGet, "/api/exam")
	if status != http.StatusOK {
		t.Fatalf("GET /api/exam got %d %q", status, body)
	}
	var state examState
	if err := json.Unmarshal([]byte(body), &state); err != nil {
		t.Fatal(err)
	}
	return state
}
