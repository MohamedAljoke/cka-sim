package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestSolvingATaskRunsItsSolutionAndAnswersEvenAFailedOne(t *testing.T) {
	l := &lab{solve: labs.ScriptResult{Output: "image not found", ExitCode: 1}}

	got, err := newSolve(l).Execute(context.Background(), owner, "wl-scale")

	if err != nil || got.ExitCode != 1 || got.Output != "image not found" || !slices.Equal(l.ran, []string{"solve"}) {
		t.Errorf("got %+v, %v; ran %v", got, err, l.ran)
	}
}

func TestSolvingATaskOnALabThatIsNotReadyStaysNotReady(t *testing.T) {
	l := &lab{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}

	_, err := newSolve(l).Execute(context.Background(), owner, "wl-scale")

	if !errors.Is(err, labs.ErrNotReady) {
		t.Errorf("got %v", err)
	}
}

func TestSolvingATaskOnALabThatDoesNotAnswerIsUnavailable(t *testing.T) {
	l := &lab{err: errors.New("connection refused")}

	_, err := newSolve(l).Execute(context.Background(), owner, "wl-scale")

	if !errors.Is(err, labs.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestSolvingAnUnknownTaskIsNotFound(t *testing.T) {
	l := &lab{}

	_, err := newSolve(l).Execute(context.Background(), owner, "net-svc")

	if !errors.Is(err, tasks.ErrNotFound) || len(l.ran) != 0 {
		t.Errorf("got %v, ran %v", err, l.ran)
	}
}

func newSolve(r ScriptRunner) *SolveTask {
	return NewSolveTask(fakeRepository{tasks: []tasks.Task{fair}}, r, newClusters())
}
