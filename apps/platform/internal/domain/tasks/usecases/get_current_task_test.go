package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestNoTaskIsCurrentBeforeASetup(t *testing.T) {
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup}}, &fakeRunner{}, &fakeAttempts{}, time.Now)

	_, err := uc.Current.Execute(context.Background(), owner)

	if !errors.Is(err, tasks.ErrNoCurrentTask) {
		t.Errorf("got %v, want ErrNoCurrentTask", err)
	}
}

func TestTheTaskSetUpLastIsCurrent(t *testing.T) {
	other := tasks.Task{ID: "nw-dns", Host: "cka-sim-control-plane", Setup: "true\n"}
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup, other}}, &fakeRunner{}, &fakeAttempts{}, time.Now)
	uc.SetUp.Execute(context.Background(), owner, "wl-scale")
	uc.SetUp.Execute(context.Background(), owner, "nw-dns")

	got, err := uc.Current.Execute(context.Background(), owner)

	if err != nil || got.Task.ID != "nw-dns" {
		t.Errorf("Execute = %q, %v", got.Task.ID, err)
	}
}

func TestACurrentTaskIsPreparingUntilItsSetupEnds(t *testing.T) {
	c := newClusters()
	c.of(owner).enter("wl-scale")

	got, err := NewGetCurrentTask(fakeRepository{tasks: []tasks.Task{withSetup}}, c).Execute(context.Background(), owner)

	if err != nil || got.Task.ID != "wl-scale" || got.Setup != tasks.SetupPreparing {
		t.Errorf("Execute = %q %q, %v", got.Task.ID, got.Setup, err)
	}
}

func TestACurrentTaskStaysPreparingUntilItsLastSetupEnds(t *testing.T) {
	c := newClusters()
	c.of(owner).enter("wl-scale")
	c.of(owner).enter("wl-scale")
	c.of(owner).settle("wl-scale", tasks.SetupReady)
	uc := NewGetCurrentTask(fakeRepository{tasks: []tasks.Task{withSetup}}, c)

	running, _ := uc.Execute(context.Background(), owner)
	c.of(owner).settle("wl-scale", tasks.SetupFailed)
	ended, _ := uc.Execute(context.Background(), owner)

	if running.Setup != tasks.SetupPreparing || ended.Setup != tasks.SetupFailed {
		t.Errorf("while one runs %q, after both %q", running.Setup, ended.Setup)
	}
}

func TestACurrentTaskWhoseSetupPassedIsReady(t *testing.T) {
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup}}, &fakeRunner{}, &fakeAttempts{}, time.Now)
	uc.SetUp.Execute(context.Background(), owner, "wl-scale")

	got, err := uc.Current.Execute(context.Background(), owner)

	if err != nil || got.Setup != tasks.SetupReady {
		t.Errorf("Execute = %q, %v", got.Setup, err)
	}
}

func TestACurrentTaskWhoseSetupFailedSaysSo(t *testing.T) {
	for name, r := range map[string]*fakeRunner{
		"exit code": {result: labs.ScriptResult{ExitCode: 1}},
		"lab down":  {err: errors.New("dial tcp: connection refused")},
	} {
		uc := New(fakeRepository{tasks: []tasks.Task{withSetup}}, r, &fakeAttempts{}, time.Now)
		uc.SetUp.Execute(context.Background(), owner, "wl-scale")

		got, err := uc.Current.Execute(context.Background(), owner)

		if err != nil || got.Setup != tasks.SetupFailed {
			t.Errorf("%s: Execute = %q, %v", name, got.Setup, err)
		}
	}
}

func TestEachOwnerHasTheirOwnCurrentTask(t *testing.T) {
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup}}, &fakeRunner{}, &fakeAttempts{}, time.Now)
	uc.SetUp.Execute(context.Background(), owner, "wl-scale")

	_, err := uc.Current.Execute(context.Background(), "bob")

	if !errors.Is(err, tasks.ErrNoCurrentTask) {
		t.Errorf("got %v, want ErrNoCurrentTask", err)
	}
}

func TestTidyingTheCurrentTaskLeavesIt(t *testing.T) {
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup}}, &fakeRunner{}, &fakeAttempts{}, time.Now)
	uc.SetUp.Execute(context.Background(), owner, "wl-scale")
	uc.Tidy.Execute(context.Background(), owner, "wl-scale")

	_, err := uc.Current.Execute(context.Background(), owner)

	if !errors.Is(err, tasks.ErrNoCurrentTask) {
		t.Errorf("got %v, want ErrNoCurrentTask", err)
	}
}

func TestTidyingAnotherTaskKeepsTheCurrentOne(t *testing.T) {
	other := tasks.Task{ID: "nw-dns", Host: "cka-sim-control-plane", Setup: "true\n"}
	uc := New(fakeRepository{tasks: []tasks.Task{withSetup, other}}, &fakeRunner{}, &fakeAttempts{}, time.Now)
	uc.SetUp.Execute(context.Background(), owner, "wl-scale")
	uc.Tidy.Execute(context.Background(), owner, "nw-dns")

	got, err := uc.Current.Execute(context.Background(), owner)

	if err != nil || got.Task.ID != "wl-scale" {
		t.Errorf("Execute = %q, %v", got.Task.ID, err)
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
