package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

var fair = tasks.Task{ID: "wl-scale", Host: "cka-sim-control-plane", Setup: "set up", Check: "check", Solve: "solve"}

func TestAFairTaskPassesItsSelftest(t *testing.T) {
	l := &lab{before: "FAIL 2 web has 4 replicas\n", after: "PASS 2 web has 4 replicas\n"}

	err := newSelftest(l).Execute(context.Background(), owner, "wl-scale")

	if err != nil || !slices.Equal(l.ran, []string{"set up", "check", "solve", "check"}) {
		t.Errorf("got %v, ran %v", err, l.ran)
	}
}

func TestATaskThatEarnsPointsBeforeTheSolutionFails(t *testing.T) {
	l := &lab{before: "PASS 1 Pod web exists\nFAIL 1 Pod web is Running\n"}

	err := newSelftest(l).Execute(context.Background(), owner, "wl-scale")

	if err == nil || err.Error() != `wl-scale earns 1/2 before the solution: "Pod web exists"` || slices.Contains(l.ran, "solve") {
		t.Errorf("got %v, ran %v", err, l.ran)
	}
}

func TestASolutionThatMissesACheckFails(t *testing.T) {
	l := &lab{before: "FAIL 1 Pod web exists\nFAIL 1 Pod web is Running\n", after: "PASS 1 Pod web exists\nFAIL 1 Pod web is Running\n"}

	err := newSelftest(l).Execute(context.Background(), owner, "wl-scale")

	if err == nil || err.Error() != `wl-scale earns 1/2 after the solution: "Pod web is Running"` {
		t.Errorf("got %v", err)
	}
}

func TestACheckThatPrintsNoChecksFails(t *testing.T) {
	l := &lab{before: "bash: jq: command not found\n"}

	err := newSelftest(l).Execute(context.Background(), owner, "wl-scale")

	if err == nil || !strings.Contains(err.Error(), "prints no checks") || !strings.Contains(err.Error(), "jq") {
		t.Errorf("got %v", err)
	}
}

func TestAFailedSetupFailsTheSelftest(t *testing.T) {
	l := &lab{setUp: labs.ScriptResult{Output: "no such node", ExitCode: 1}}

	err := newSelftest(l).Execute(context.Background(), owner, "wl-scale")

	if err == nil || !strings.Contains(err.Error(), "no such node") || len(l.ran) != 1 {
		t.Errorf("got %v, ran %v", err, l.ran)
	}
}

func TestAFailedSolutionFailsTheSelftest(t *testing.T) {
	l := &lab{before: "FAIL 1 Pod web exists\n", solve: labs.ScriptResult{Output: "image not found", ExitCode: 1}}

	err := newSelftest(l).Execute(context.Background(), owner, "wl-scale")

	if err == nil || !strings.Contains(err.Error(), "solve wl-scale") || !strings.Contains(err.Error(), "image not found") {
		t.Errorf("got %v", err)
	}
}

func TestASelftestOnALabThatIsNotReadyStaysNotReady(t *testing.T) {
	l := &lab{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}

	err := newSelftest(l).Execute(context.Background(), owner, "wl-scale")

	if !errors.Is(err, labs.ErrNotReady) {
		t.Errorf("got %v", err)
	}
}

func TestSelftestingAnUnknownTaskIsNotFound(t *testing.T) {
	l := &lab{}

	err := newSelftest(l).Execute(context.Background(), owner, "net-svc")

	if !errors.Is(err, tasks.ErrNotFound) || len(l.ran) != 0 {
		t.Errorf("got %v, ran %v", err, l.ran)
	}
}

func newSelftest(r ScriptRunner) *SelftestTask {
	repo, c := fakeRepository{tasks: []tasks.Task{fair}}, newClusters()
	start := NewStartTask(repo, NewHealCluster(repo, r, c), NewSetUpTask(repo, r, c))

	return NewSelftestTask(start, NewCheckTask(repo, r), NewSolveTask(repo, r, c))
}

// lab answers the fair task's scripts, and its check passes once the solution ran.
type lab struct {
	setUp, solve  labs.ScriptResult
	before, after string
	err           error
	solved        bool
	ran           []string
}

func (l *lab) RunScript(_ context.Context, _, _, script string) (labs.ScriptResult, error) {
	l.ran = append(l.ran, script)

	switch {
	case script == fair.Solve:
		l.solved = true
		return l.solve, l.err
	case script == fair.Check && l.solved:
		return labs.ScriptResult{Output: l.after}, l.err
	case script == fair.Check:
		return labs.ScriptResult{Output: l.before}, l.err
	}

	return l.setUp, l.err
}
