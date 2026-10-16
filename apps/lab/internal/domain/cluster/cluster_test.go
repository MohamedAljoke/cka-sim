package cluster

import (
	"errors"
	"testing"
	"time"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

var now = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)

func TestANewClusterIsAbsent(t *testing.T) {
	if got := New().Status(); got.State != Absent {
		t.Errorf("state = %s", got.State)
	}
}

func TestCreatingGoesThroughItsStepsToReady(t *testing.T) {
	c := New()

	start, err := c.BeginCreate(now)
	if !start || err != nil {
		t.Fatalf("BeginCreate = %v, %v", start, err)
	}
	c.Progress("Preparing nodes")
	if got := c.Status(); got.State != Creating || got.Step != "Preparing nodes" || !got.Started.Equal(now) {
		t.Fatalf("while creating: %+v", got)
	}

	c.Created(now.Add(time.Minute))

	got := c.Status()
	if got.State != Ready || got.Step != "" || !got.Started.Equal(now) || !got.Ready.Equal(now.Add(time.Minute)) {
		t.Errorf("after creating: %+v", got)
	}
}

func TestBeginCreateAsksForOneBuildOnly(t *testing.T) {
	for _, state := range []labv1.ClusterState{Creating, Ready} {
		c := clusterIn(state)

		start, err := c.BeginCreate(now.Add(time.Hour))

		if start || err != nil || c.Status().State != state {
			t.Errorf("%s: BeginCreate = %v, %v; state %s", state, start, err, c.Status().State)
		}
	}
}

func TestAFailedClusterCanBeCreatedAgain(t *testing.T) {
	c := clusterIn(Failed)

	start, err := c.BeginCreate(now)

	if !start || err != nil || c.Status().State != Creating || c.Status().Error != "" {
		t.Errorf("BeginCreate = %v, %v; status %+v", start, err, c.Status())
	}
}

func TestACreateFailureSaysWhereTheCauseIs(t *testing.T) {
	c := clusterIn(Creating)

	c.CreateFailed()

	if got := c.Status(); got.State != Failed || got.Error == "" {
		t.Errorf("status %+v", got)
	}
}

func TestDeletingEndsAbsent(t *testing.T) {
	for _, state := range []labv1.ClusterState{Ready, Failed} {
		c := clusterIn(state)

		start, err := c.BeginDelete()
		if !start || err != nil || c.Status().State != Deleting {
			t.Fatalf("%s: BeginDelete = %v, %v; state %s", state, start, err, c.Status().State)
		}
		c.Deleted()

		if got := c.Status(); got.State != Absent {
			t.Errorf("%s: after deleting %+v", state, got)
		}
	}
}

func TestBeginDeleteHasNothingToDoWhenAbsentOrDeleting(t *testing.T) {
	for _, state := range []labv1.ClusterState{Absent, Deleting} {
		start, err := clusterIn(state).BeginDelete()

		if start || err != nil {
			t.Errorf("%s: BeginDelete = %v, %v", state, start, err)
		}
	}
}

func TestAClusterIsBusyWhileTheOtherChangeRuns(t *testing.T) {
	if _, err := clusterIn(Deleting).BeginCreate(now); !errors.Is(err, ErrBusy) {
		t.Errorf("create while deleting: %v", err)
	}
	if _, err := clusterIn(Creating).BeginDelete(); !errors.Is(err, ErrBusy) {
		t.Errorf("delete while creating: %v", err)
	}
}

func TestALateResultCannotOverwriteAnotherState(t *testing.T) {
	c := clusterIn(Deleting)

	c.Created(now)
	c.CreateFailed()
	c.Progress("Preparing nodes")

	if got := c.Status(); got.State != Deleting || got.Step != "" {
		t.Errorf("status %+v", got)
	}
}

func clusterIn(state labv1.ClusterState) *Cluster {
	c := New()
	switch state {
	case Creating:
		c.BeginCreate(now)
	case Ready:
		c.BeginCreate(now)
		c.Created(now)
	case Failed:
		c.BeginCreate(now)
		c.CreateFailed()
	case Deleting:
		c.BeginCreate(now)
		c.Created(now)
		c.BeginDelete()
	}
	return c
}
