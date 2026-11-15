package usecases

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestMovingToATaskTidiesTheOneTheOwnerIsOnFirst(t *testing.T) {
	r := &fakeRunner{}
	uc := New(breakers, r, &fakeAttempts{}, time.Now)
	uc.MoveTo.Execute(context.Background(), owner, "tr-kubelet")

	task, started, err := uc.MoveTo.Execute(context.Background(), owner, "wl-scale")

	want := []string{"break kubelet", "fix kubelet", TidyScript("tr-kubelet"), "fresh_ns wl-scale\n"}
	if err != nil || !started.Passed() || task.ID != "wl-scale" || !slices.Equal(r.ran, want) {
		t.Errorf("Execute = %+v, %+v, %v; ran %v", task, started, err, r.ran)
	}
	if current, _ := uc.Current.Execute(context.Background(), owner); current.Task.ID != "wl-scale" {
		t.Errorf("current %q", current.Task.ID)
	}
}

func TestMovingToTheTaskTheOwnerIsOnSetsItUpAgainWithoutATidy(t *testing.T) {
	r := &fakeRunner{}
	uc := New(breakers, r, &fakeAttempts{}, time.Now)
	uc.MoveTo.Execute(context.Background(), owner, "wl-scale")

	uc.MoveTo.Execute(context.Background(), owner, "wl-scale")

	if !slices.Equal(r.ran, []string{"fresh_ns wl-scale\n", "fresh_ns wl-scale\n"}) {
		t.Errorf("ran %v", r.ran)
	}
}

func TestMovingToAnUnknownTaskTidiesNothing(t *testing.T) {
	r := &fakeRunner{}
	uc := New(breakers, r, &fakeAttempts{}, time.Now)
	uc.MoveTo.Execute(context.Background(), owner, "tr-kubelet")

	_, _, err := uc.MoveTo.Execute(context.Background(), owner, "no-such-task")

	if !errors.Is(err, tasks.ErrNotFound) || len(r.ran) != 1 {
		t.Errorf("got %v, ran %v", err, r.ran)
	}
}

func TestAFailedTidyLeavesTheResetToTheStartsHeal(t *testing.T) {
	r := &fakeRunner{results: map[string]labs.ScriptResult{"fix kubelet": {ExitCode: 1}}}
	uc := New(breakers, r, &fakeAttempts{}, time.Now)
	uc.MoveTo.Execute(context.Background(), owner, "tr-kubelet")

	_, started, err := uc.MoveTo.Execute(context.Background(), owner, "wl-scale")

	if err != nil || started.Passed() || !slices.Equal(r.ran, []string{"break kubelet", "fix kubelet", "fix kubelet"}) {
		t.Errorf("Execute = %+v, %v; ran %v", started, err, r.ran)
	}
}

func TestMovingOnALabThatIsNotReadyStaysNotReady(t *testing.T) {
	r := &fakeRunner{}
	uc := New(breakers, r, &fakeAttempts{}, time.Now)
	uc.MoveTo.Execute(context.Background(), owner, "wl-scale")
	r.err = labs.ErrNotReady

	_, _, err := uc.MoveTo.Execute(context.Background(), owner, "tr-etcd")

	if !errors.Is(err, labs.ErrNotReady) {
		t.Errorf("got %v", err)
	}
}
