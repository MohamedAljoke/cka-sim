package local

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
)

func TestCreateAsksForMakeUpWhenNoLabAnswers(t *testing.T) {
	_, err := (&Provider{Agent: labagent.Client{URL: "http://127.0.0.1:1"}}).Create(context.Background(), "a")

	if err == nil || !strings.Contains(err.Error(), "run make up") {
		t.Errorf("err = %v, want one that says to run make up", err)
	}
}

func TestCreateProbesTheClusterAndEndsLeftoverShells(t *testing.T) {
	agent := startFakeAgent(t)

	box, err := (&Provider{Agent: labagent.Client{URL: agent.url}}).Create(context.Background(), "a")

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

func TestASecondAccountWaitsUntilTheFirstEndsItsLab(t *testing.T) {
	agent := startFakeAgent(t)
	p := &Provider{Agent: labagent.Client{URL: agent.url}}
	box, err := p.Create(context.Background(), "a")
	if err != nil {
		t.Fatal(err)
	}
	agent.shellsEnded = false

	_, refused := p.Create(context.Background(), "b")
	if !errors.Is(refused, ErrInUse) {
		t.Fatalf("second account: %v, want ErrInUse", refused)
	}
	if agent.shellsEnded {
		t.Error("the refused account closed the first one's shells")
	}
	if _, err := p.Create(context.Background(), "a"); err != nil {
		t.Errorf("the owner again: %v", err)
	}
	if err := p.Destroy(context.Background(), box); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Create(context.Background(), "b"); err != nil {
		t.Errorf("after the first ended: %v", err)
	}
}

func TestAFailedCreateLeavesTheLabFree(t *testing.T) {
	p := &Provider{Agent: labagent.Client{URL: "http://127.0.0.1:1"}}
	if _, err := p.Create(context.Background(), "a"); err == nil {
		t.Fatal("create succeeded with no lab")
	}

	p.Agent.URL = startFakeAgent(t).url
	if _, err := p.Create(context.Background(), "b"); err != nil {
		t.Errorf("b after a's failed create: %v", err)
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
