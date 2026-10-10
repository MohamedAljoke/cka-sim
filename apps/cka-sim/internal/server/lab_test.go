package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
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

func getLab(t *testing.T, srv *httptest.Server) sandbox.Status {
	status, body := call(t, srv, http.MethodGet, "/api/lab")
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
