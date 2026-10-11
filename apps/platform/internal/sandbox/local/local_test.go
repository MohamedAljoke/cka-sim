package local

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
)

func TestCreateAsksForMakeUpWhenNoLabAnswers(t *testing.T) {
	_, err := Provider{Agent: labagent.Client{URL: "http://127.0.0.1:1"}}.Create(context.Background())

	if err == nil || !strings.Contains(err.Error(), "run make up") {
		t.Errorf("err = %v, want one that says to run make up", err)
	}
}

func TestCreateProbesTheClusterAndEndsLeftoverShells(t *testing.T) {
	agent := startFakeAgent(t)

	box, err := Provider{Agent: labagent.Client{URL: agent.url}}.Create(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if box.Runner == nil || box.Shells == nil {
		t.Errorf("box = %+v, want the agent as runner and shells", box)
	}
	if agent.ranOn != labagent.ControlPlaneNode {
		t.Errorf("probe ran on %q, want %q", agent.ranOn, labagent.ControlPlaneNode)
	}
	if !agent.shellsEnded {
		t.Error("leftover shells were not ended")
	}
}

type fakeAgent struct {
	url         string
	ranOn       string
	shellsEnded bool
}

func startFakeAgent(t *testing.T) *fakeAgent {
	a := &fakeAgent{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("POST /run", func(w http.ResponseWriter, r *http.Request) {
		a.ranOn = r.URL.Query().Get("host")
		json.NewEncoder(w).Encode(labagent.RunResult{Output: "ok"})
	})
	mux.HandleFunc("POST /shells/end", func(w http.ResponseWriter, r *http.Request) {
		a.shellsEnded = true
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	a.url = srv.URL
	return a
}
