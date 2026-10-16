package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

func TestDeleteClusterReturnsAtOnceAndDeletesInTheBackground(t *testing.T) {
	p := newFakeProvisioner()
	c := readyCluster(t, p)

	got, err := NewDeleteCluster(c, p).Execute(context.Background())
	if err != nil || got.State != cluster.Deleting {
		t.Fatalf("Execute = %+v, %v", got, err)
	}

	p.deleted <- nil
	waitFor(t, c, func(s cluster.Status) bool { return s.State == cluster.Absent })
}

func TestDeletingAnAbsentClusterDoesNothing(t *testing.T) {
	p := newFakeProvisioner()

	got, err := NewDeleteCluster(cluster.New(), p).Execute(context.Background())

	if err != nil || got.State != cluster.Absent || p.deletes.Load() != 0 {
		t.Errorf("Execute = %+v, %v; Delete ran %d times", got, err, p.deletes.Load())
	}
}

func TestAFailedDeleteLeavesTheClusterFailed(t *testing.T) {
	p := newFakeProvisioner()
	c := readyCluster(t, p)

	NewDeleteCluster(c, p).Execute(context.Background())
	p.deleted <- errors.New("docker: daemon gone")

	waitFor(t, c, func(s cluster.Status) bool { return s.State == cluster.Failed })
}

func TestDeleteClusterIsBusyWhileCreating(t *testing.T) {
	p := newFakeProvisioner()
	c := cluster.New()
	NewCreateCluster(c, p, clock).Execute(context.Background())

	_, err := NewDeleteCluster(c, p).Execute(context.Background())

	if !errors.Is(err, cluster.ErrBusy) {
		t.Errorf("got %v, want ErrBusy", err)
	}
	p.created <- nil
}

func readyCluster(t *testing.T, p *fakeProvisioner) *cluster.Cluster {
	t.Helper()
	c := cluster.New()
	NewCreateCluster(c, p, clock).Execute(context.Background())
	p.created <- nil
	waitFor(t, c, func(s cluster.Status) bool { return s.State == cluster.Ready })
	return c
}
