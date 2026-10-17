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

var withCheck = tasks.Task{ID: "wl-scale", Host: "cka-sim-control-plane", Check: "check 2 \"web has 4 replicas\" true\n"}

func TestCheckTaskGradesTheCheckItRunsOnTheTasksHost(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "PASS 2 web has 4 replicas\nFAIL 1 web is Ready\n"}}

	got, err := NewCheckTask(fakeRepository{tasks: []tasks.Task{withCheck}}, r).Execute(context.Background(), "wl-scale")

	if err != nil || got.Grade.Earned != 2 || got.Grade.Total != 3 || len(got.Grade.Checks) != 2 {
		t.Fatalf("Execute = %+v, %v", got, err)
	}
	if r.host != "cka-sim-control-plane" || r.script != withCheck.Check {
		t.Errorf("ran %q on %q", r.script, r.host)
	}
}

func TestACheckThatPrintsNoResultKeepsItsOutput(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "bash: jq: command not found\n", ExitCode: 127}}

	got, err := NewCheckTask(fakeRepository{tasks: []tasks.Task{withCheck}}, r).Execute(context.Background(), "wl-scale")

	if err != nil || len(got.Grade.Checks) != 0 || got.Script.ExitCode != 127 || !strings.Contains(got.Script.Output, "jq") {
		t.Errorf("Execute = %+v, %v", got, err)
	}
}

func TestACheckStopsWithTheRequestAndHasATimeout(t *testing.T) {
	r := &fakeRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	NewCheckTask(fakeRepository{tasks: []tasks.Task{withCheck}}, r).Execute(ctx, "wl-scale")

	if r.ctxErr == nil || !r.hasDeadline {
		t.Errorf("the check ran with ctx error %v, deadline %v", r.ctxErr, r.hasDeadline)
	}
}

func TestCheckTaskOfAnUnknownTaskIsNotFound(t *testing.T) {
	r := &fakeRunner{}

	_, err := NewCheckTask(fakeRepository{tasks: []tasks.Task{withCheck}}, r).Execute(context.Background(), "net-svc")

	if !errors.Is(err, tasks.ErrNotFound) || r.host != "" {
		t.Errorf("got %v, ran on %q", err, r.host)
	}
}

func TestCheckTaskOnALabThatIsNotReadyStaysNotReady(t *testing.T) {
	r := &fakeRunner{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}

	_, err := NewCheckTask(fakeRepository{tasks: []tasks.Task{withCheck}}, r).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, labs.ErrNotReady) || errors.Is(err, labs.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestCheckTaskOnALabThatDoesNotAnswerIsUnavailableAndKeepsTheCause(t *testing.T) {
	r := &fakeRunner{err: errors.New("dial tcp 127.0.0.1:7171: connection refused")}

	_, err := NewCheckTask(fakeRepository{tasks: []tasks.Task{withCheck}}, r).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, labs.ErrUnavailable) || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("got %v", err)
	}
}

func TestCheckTaskIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	_, err := NewCheckTask(fakeRepository{err: errors.New("disk gone")}, &fakeRunner{}).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, tasks.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}
