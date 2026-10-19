package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestNoTaskIsCurrentBeforeASetup(t *testing.T) {
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup}}, &fakeRunner{})

	_, err := uc.Current.Execute(context.Background(), owner)

	if !errors.Is(err, tasks.ErrNoCurrentTask) {
		t.Errorf("got %v, want ErrNoCurrentTask", err)
	}
}

func TestTheTaskSetUpLastIsCurrent(t *testing.T) {
	other := tasks.Task{ID: "nw-dns", Host: "cka-sim-control-plane", Setup: "true\n"}
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup, other}}, &fakeRunner{})
	uc.SetUp.Execute(context.Background(), owner, "wl-scale")
	uc.SetUp.Execute(context.Background(), owner, "nw-dns")

	got, err := uc.Current.Execute(context.Background(), owner)

	if err != nil || got.ID != "nw-dns" {
		t.Errorf("Execute = %q, %v", got.ID, err)
	}
}

func TestEachOwnerHasTheirOwnCurrentTask(t *testing.T) {
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup}}, &fakeRunner{})
	uc.SetUp.Execute(context.Background(), owner, "wl-scale")

	_, err := uc.Current.Execute(context.Background(), "bob")

	if !errors.Is(err, tasks.ErrNoCurrentTask) {
		t.Errorf("got %v, want ErrNoCurrentTask", err)
	}
}

func TestTidyingTheCurrentTaskLeavesIt(t *testing.T) {
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup}}, &fakeRunner{})
	uc.SetUp.Execute(context.Background(), owner, "wl-scale")
	uc.Tidy.Execute(context.Background(), owner, "wl-scale")

	_, err := uc.Current.Execute(context.Background(), owner)

	if !errors.Is(err, tasks.ErrNoCurrentTask) {
		t.Errorf("got %v, want ErrNoCurrentTask", err)
	}
}

func TestTidyingAnotherTaskKeepsTheCurrentOne(t *testing.T) {
	other := tasks.Task{ID: "nw-dns", Host: "cka-sim-control-plane", Setup: "true\n"}
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup, other}}, &fakeRunner{})
	uc.SetUp.Execute(context.Background(), owner, "wl-scale")
	uc.Tidy.Execute(context.Background(), owner, "nw-dns")

	got, err := uc.Current.Execute(context.Background(), owner)

	if err != nil || got.ID != "wl-scale" {
		t.Errorf("Execute = %q, %v", got.ID, err)
	}
}

func TestTheCurrentTaskIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	c := newClusters()
	c.of(owner).enter("wl-scale")
	cause := errors.New("disk gone")

	_, err := NewGetCurrentTask(fakeRepository{err: cause}, c).Execute(context.Background(), owner)

	if !errors.Is(err, tasks.ErrUnavailable) || !errors.Is(err, cause) {
		t.Errorf("got %v", err)
	}
}
