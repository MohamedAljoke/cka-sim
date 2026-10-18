package usecases

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestStartHealsWhatAnEarlierTaskBrokeThenSetsUp(t *testing.T) {
	r := &fakeRunner{}
	start := newStart(breakers, r)

	start.Execute(context.Background(), "tr-kubelet")
	got, err := start.Execute(context.Background(), "wl-scale")

	if err != nil || !got.Passed() || !slices.Equal(r.ran, []string{"break kubelet", "fix kubelet", "fresh_ns wl-scale\n"}) {
		t.Errorf("Execute = %+v, %v; ran %v", got, err, r.ran)
	}
}

func TestAFailedResetAnswersInPlaceOfTheSetup(t *testing.T) {
	r := &fakeRunner{results: map[string]labs.ScriptResult{"fix kubelet": {ExitCode: 1}}}
	start := newStart(breakers, r)
	start.Execute(context.Background(), "tr-kubelet")

	got, err := start.Execute(context.Background(), "wl-scale")

	if err != nil || got.Passed() || slices.Contains(r.ran, "fresh_ns wl-scale\n") {
		t.Errorf("Execute = %+v, %v; ran %v", got, err, r.ran)
	}
}

func TestStartingAnUnknownTaskHealsNothing(t *testing.T) {
	r := &fakeRunner{}
	start := newStart(breakers, r)
	start.Execute(context.Background(), "tr-kubelet")

	_, err := start.Execute(context.Background(), "net-svc")

	if !errors.Is(err, tasks.ErrNotFound) || len(r.ran) != 1 {
		t.Errorf("got %v, ran %v", err, r.ran)
	}
}

func TestStartIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	_, err := newStart(fakeRepository{err: errors.New("disk gone")}, &fakeRunner{}).Execute(context.Background(), "wl-scale")

	if !errors.Is(err, tasks.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func newStart(repo TaskRepository, r ScriptRunner) *StartTask {
	c := newCluster()
	return NewStartTask(repo, NewHealCluster(repo, r, c), NewSetUpTask(repo, r, c))
}
