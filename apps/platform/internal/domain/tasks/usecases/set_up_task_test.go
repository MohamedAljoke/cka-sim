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

func TestSetUpTaskRunsTheSetupOnTheTasksHost(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "namespace/wl-scale created\n"}}

	got, err := NewSetUpTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r, newCluster()).Execute(context.Background(), "wl-scale")

	if err != nil || !got.Passed() || got.Output != "namespace/wl-scale created\n" {
		t.Fatalf("Execute = %+v, %v", got, err)
	}
	if r.host != "cka-sim-control-plane" || r.script != "fresh_ns wl-scale\n" {
		t.Errorf("ran %q on %q", r.script, r.host)
	}
}

func TestASetupTouchesItsTaskEvenWhenItFails(t *testing.T) {
	r, c := &fakeRunner{result: labs.ScriptResult{ExitCode: 1}}, newCluster()

	NewSetUpTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r, c).Execute(context.Background(), "wl-scale")

	if !c.isTouched("wl-scale") {
		t.Error("the task isn't touched")
	}
}

func TestAFailedSetupIsAnAnswerNotAnError(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "error: no such CRD\n", ExitCode: 1}}

	got, err := NewSetUpTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r, newCluster()).Execute(context.Background(), "wl-scale")

	if err != nil || got.Passed() || got.ExitCode != 1 {
		t.Errorf("Execute = %+v, %v", got, err)
	}
}

func TestASetupOutlivesTheRequestButNotItsTimeout(t *testing.T) {
	r := &fakeRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	NewSetUpTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r, newCluster()).Execute(ctx, "wl-scale")

	if r.ctxErr != nil || !r.hasDeadline {
		t.Errorf("the setup ran with ctx error %v, deadline %v", r.ctxErr, r.hasDeadline)
	}
}

func TestSetUpTaskOfAnUnknownTaskIsNotFound(t *testing.T) {
	r := &fakeRunner{}

	_, err := NewSetUpTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r, newCluster()).Execute(context.Background(), "net-svc")

	if !errors.Is(err, tasks.ErrNotFound) || r.host != "" {
		t.Errorf("got %v, ran on %q", err, r.host)
	}
}

func TestSetUpTaskOnALabThatIsNotReadyStaysNotReady(t *testing.T) {
	r := &fakeRunner{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}

	_, err := NewSetUpTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r, newCluster()).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, labs.ErrNotReady) || errors.Is(err, labs.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestSetUpTaskOnALabThatDoesNotAnswerIsUnavailableAndKeepsTheCause(t *testing.T) {
	r := &fakeRunner{err: errors.New("dial tcp 127.0.0.1:7171: connection refused")}

	_, err := NewSetUpTask(fakeRepository{tasks: []tasks.Task{withSetup}}, r, newCluster()).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, labs.ErrUnavailable) || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("got %v", err)
	}
}

func TestSetUpTaskIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	_, err := NewSetUpTask(fakeRepository{err: errors.New("disk gone")}, &fakeRunner{}, newCluster()).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, tasks.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}
