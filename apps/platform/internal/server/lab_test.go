package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

func TestStartAndEndALab(t *testing.T) {
	srv, _ := newAPIWithoutLab(t, &fakeRunner{})

	if got := getLab(t, srv); got.State != sandbox.None || got.Provider != "Fake" {
		t.Fatalf("before start: %+v, want none from Fake", got)
	}
	if status, body := call(t, srv, http.MethodPost, "/api/lab"); status != http.StatusAccepted {
		t.Fatalf("start got %d %q", status, body)
	}
	waitLab(t, srv, sandbox.Ready)
	if status, body := call(t, srv, http.MethodDelete, "/api/lab"); status != http.StatusNoContent {
		t.Fatalf("end got %d %q", status, body)
	}
	if got := getLab(t, srv); got.State != sandbox.None {
		t.Errorf("after end: %+v, want none", got)
	}
}

func TestPracticeStartsTheLab(t *testing.T) {
	r := &fakeRunner{}
	srv, _ := newAPIWithoutLab(t, r)

	status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start")

	if status != http.StatusNoContent {
		t.Errorf("got %d %q, want 204", status, body)
	}
	if got := getLab(t, srv); got.State != sandbox.Ready {
		t.Errorf("lab %+v, want ready", got)
	}
	if ran := r.ran(); len(ran) == 0 {
		t.Error("the task was not set up")
	}
}

func TestBeginExamStartsTheLab(t *testing.T) {
	r := &fakeRunner{}
	srv, _ := newAPIWithoutLab(t, r)

	beginExam(t, srv)

	if got := getLab(t, srv); got.State != sandbox.Ready {
		t.Errorf("lab %+v, want ready", got)
	}
	if ran := r.ran(); len(ran) == 0 {
		t.Error("the exam set nothing up")
	}
}

func TestEndLabDuringAnExamIs409(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	beginExam(t, srv)

	status, body := call(t, srv, http.MethodDelete, "/api/lab")

	if status != http.StatusConflict {
		t.Errorf("got %d %q, want 409", status, body)
	}
}

func TestEachAccountHasItsOwnLab(t *testing.T) {
	srv, _ := newAPIWithoutLab(t, &fakeRunner{})

	if status, body := requestAs(t, srv, "a", http.MethodPost, "/api/tasks/wl-scale/start", ""); status != http.StatusNoContent {
		t.Fatalf("a's start got %d %q", status, body)
	}

	if got := getLabAs(t, srv, "b"); got.State != sandbox.None {
		t.Errorf("b's lab is %s, want none", got.State)
	}
	if status, body := requestAs(t, srv, "b", http.MethodPost, "/api/tasks/wl-scale/start", ""); status != http.StatusNoContent {
		t.Fatalf("b's start got %d %q", status, body)
	}
	if status, body := requestAs(t, srv, "a", http.MethodDelete, "/api/lab", ""); status != http.StatusNoContent {
		t.Fatalf("a's end got %d %q", status, body)
	}
	if a, b := getLabAs(t, srv, "a"), getLabAs(t, srv, "b"); a.State != sandbox.None || b.State != sandbox.Ready {
		t.Errorf("after a ends: a %s, b %s; want none and ready", a.State, b.State)
	}
}

func TestOneAccountsScriptDoesNotBlockAnother(t *testing.T) {
	r := &fakeRunner{started: make(chan struct{}), release: make(chan struct{})}
	srv, _ := newAPIWithoutLab(t, r)
	done := make(chan struct{})
	go func() {
		defer close(done)
		requestAs(t, srv, "a", http.MethodPost, "/api/tasks/wl-scale/start", "")
	}()
	<-r.started

	status, body := requestAs(t, srv, "b", http.MethodPost, "/api/tasks/wl-scale/start", "")

	close(r.release)
	<-done
	if status != http.StatusNoContent {
		t.Errorf("b's start got %d %q while a's script ran, want 204", status, body)
	}
}

func TestAnExamBelongsToItsAccount(t *testing.T) {
	srv, _ := newAPIWithoutLab(t, &fakeRunner{})
	beginExam(t, srv)

	status, body := requestAs(t, srv, "b", http.MethodGet, "/api/exam", "")

	if status != http.StatusNotFound {
		t.Errorf("b's exam got %d %q, want 404", status, body)
	}
	if status, body := requestAs(t, srv, "b", http.MethodPost, "/api/tasks/wl-scale/start", ""); status != http.StatusNoContent {
		t.Errorf("b's practice got %d %q during a's exam, want 204", status, body)
	}
}

func TestARequestWithoutAnAccountIsRefused(t *testing.T) {
	srv := httptest.NewServer(New(readyDesks(t, &fakeRunner{}, newFakeOpener()), Practice{}))
	t.Cleanup(srv.Close)

	status, body := call(t, srv, http.MethodGet, "/api/lab")

	if status != http.StatusUnauthorized || problemCode(t, body) != "not_logged_in" {
		t.Errorf("got %d %q, want 401 not_logged_in", status, body)
	}
}

func getLab(t *testing.T, srv *httptest.Server) sandbox.Status {
	return getLabAs(t, srv, "a")
}

func getLabAs(t *testing.T, srv *httptest.Server, account string) sandbox.Status {
	status, body := requestAs(t, srv, account, http.MethodGet, "/api/lab", "")
	if status != http.StatusOK {
		t.Fatalf("GET /api/lab got %d %q", status, body)
	}
	var state labState
	if err := json.Unmarshal([]byte(body), &state); err != nil {
		t.Fatal(err)
	}
	return state.Lab
}

func waitLab(t *testing.T, srv *httptest.Server, want sandbox.State) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if getLab(t, srv).State == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("the lab never became %s", want)
}
