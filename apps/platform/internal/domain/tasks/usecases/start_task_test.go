package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

var withSetup = tasks.Task{ID: "wl-scale", Host: "cka-sim-control-plane", Setup: "fresh_ns wl-scale\n"}

func TestStartTaskRunsTheSetupOnTheTasksHost(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "namespace/wl-scale created\n"}}

	got, err := NewStartTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r).Execute(context.Background(), "wl-scale")

	if err != nil || !got.Passed() || got.Output != "namespace/wl-scale created\n" {
		t.Fatalf("Execute = %+v, %v", got, err)
	}
	if r.host != "cka-sim-control-plane" || r.script != "fresh_ns wl-scale\n" {
		t.Errorf("ran %q on %q", r.script, r.host)
	}
}

func TestAFailedSetupIsAnAnswerNotAnError(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "error: no such CRD\n", ExitCode: 1}}

	got, err := NewStartTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r).Execute(context.Background(), "wl-scale")

	if err != nil || got.Passed() || got.ExitCode != 1 {
		t.Errorf("Execute = %+v, %v", got, err)
	}
}

func TestASetupOutlivesTheRequestButNotItsTimeout(t *testing.T) {
	r := &fakeRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	NewStartTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r).Execute(ctx, "wl-scale")

	if r.ctxErr != nil || !r.hasDeadline {
		t.Errorf("the setup ran with ctx error %v, deadline %v", r.ctxErr, r.hasDeadline)
	}
}

func TestStartTaskOfAnUnknownTaskIsNotFound(t *testing.T) {
	r := &fakeRunner{}

	_, err := NewStartTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r).Execute(context.Background(), "net-svc")

	if !errors.Is(err, tasks.ErrNotFound) || r.host != "" {
		t.Errorf("got %v, ran on %q", err, r.host)
	}
}

func TestStartTaskOnALabThatIsNotReadyStaysNotReady(t *testing.T) {
	r := &fakeRunner{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}

	_, err := NewStartTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, labs.ErrNotReady) || errors.Is(err, labs.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestStartTaskOnALabThatDoesNotAnswerIsUnavailableAndKeepsTheCause(t *testing.T) {
	r := &fakeRunner{err: errors.New("dial tcp 127.0.0.1:7171: connection refused")}

	_, err := NewStartTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, labs.ErrUnavailable) || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("got %v", err)
	}
}

func TestStartTaskIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	_, err := NewStartTask(fakeRepository{err: errors.New("disk gone")}, &fakeRunner{}).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, tasks.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}
