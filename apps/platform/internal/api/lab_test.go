package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

func TestTheLabStatusSaysWhatTheLabIsDoing(t *testing.T) {
	started := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	p := &fakeLabProvider{lab: labs.Lab{State: labs.Starting, Step: "Preparing nodes", StartedAt: started}}

	status, body := call(t, newLabAPI(p), http.MethodGet, "/api/lab")

	if status != http.StatusOK || strings.TrimSpace(body) != `{"state":"starting","step":"Preparing nodes","startedAt":"2026-10-16T09:00:00Z"}` {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestANewLabShowsOnlyItsState(t *testing.T) {
	status, body := call(t, newLabAPI(&fakeLabProvider{lab: labs.Lab{State: labs.None}}), http.MethodGet, "/api/lab")

	if status != http.StatusOK || strings.TrimSpace(body) != `{"state":"none"}` {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestStartingAndStoppingAreAccepted(t *testing.T) {
	for method, wantCalls := range map[string]string{http.MethodPost: "start", http.MethodDelete: "stop"} {
		p := &fakeLabProvider{lab: labs.Lab{State: labs.Starting}}

		status, _ := call(t, newLabAPI(p), method, "/api/lab")

		if status != http.StatusAccepted || strings.Join(p.calls, ",") != wantCalls {
			t.Errorf("%s: status %d, calls %v", method, status, p.calls)
		}
	}
}

func TestABusyLabIsAConflict(t *testing.T) {
	p := &fakeLabProvider{err: fmt.Errorf("%w: delete while creating", labs.ErrBusy)}

	status, body := call(t, newLabAPI(p), http.MethodDelete, "/api/lab")

	if status != http.StatusConflict || !strings.Contains(body, `"code":"lab_busy"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestALabThatDoesNotAnswerIsUnavailableAndHidesTheCause(t *testing.T) {
	p := &fakeLabProvider{err: errors.New("dial tcp 127.0.0.1:7171: connection refused")}

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		status, body := call(t, newLabAPI(p), method, "/api/lab")

		if status != http.StatusServiceUnavailable || !strings.Contains(body, `"code":"lab_unavailable"`) || strings.Contains(body, "7171") {
			t.Errorf("%s: status %d: %s", method, status, body)
		}
	}
}

type fakeLabProvider struct {
	lab   labs.Lab
	err   error
	calls []string
}

func (p *fakeLabProvider) Get(context.Context) (labs.Lab, error) { return p.lab, p.err }

func (p *fakeLabProvider) Start(context.Context) (labs.Lab, error) {
	p.calls = append(p.calls, "start")
	return p.lab, p.err
}

func (p *fakeLabProvider) Stop(context.Context) (labs.Lab, error) {
	p.calls = append(p.calls, "stop")
	return p.lab, p.err
}

func newLabAPI(p labusecases.LabProvider) http.Handler {
	return New(usecases.UseCases{}, labusecases.New(p))
}
