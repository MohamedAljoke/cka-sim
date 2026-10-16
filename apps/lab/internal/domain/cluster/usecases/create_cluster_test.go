package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

func TestCreateClusterReturnsAtOnceAndBuildsInTheBackground(t *testing.T) {
	p := newFakeProvisioner()
	c := cluster.New()
	uc := NewCreateCluster(c, p, clock)

	got, err := uc.Execute(context.Background())
	if err != nil || got.State != cluster.Creating || !got.Started.Equal(now) {
		t.Fatalf("Execute = %+v, %v", got, err)
	}

	p.progress <- "Preparing nodes"
	waitFor(t, c, func(s cluster.Status) bool { return s.Step == "Preparing nodes" })
	p.created <- nil
	waitFor(t, c, func(s cluster.Status) bool { return s.State == cluster.Ready })
}

func TestCreateClusterBuildsOnceWhileCreating(t *testing.T) {
	p := newFakeProvisioner()
	uc := NewCreateCluster(cluster.New(), p, clock)

	uc.Execute(context.Background())
	got, err := uc.Execute(context.Background())

	if err != nil || got.State != cluster.Creating {
		t.Fatalf("second Execute = %+v, %v", got, err)
	}
	p.created <- nil
	if calls := p.creates.Load(); calls != 1 {
		t.Errorf("Create ran %d times", calls)
	}
}

func TestCreateClusterOutlivesTheRequest(t *testing.T) {
	p := newFakeProvisioner()
	c := cluster.New()
	ctx, cancel := context.WithCancel(context.Background())

	NewCreateCluster(c, p, clock).Execute(ctx)
	cancel()
	p.created <- nil

	if err := <-p.createCtxErr; err != nil {
		t.Fatalf("the build's context ended with the request: %v", err)
	}
	waitFor(t, c, func(s cluster.Status) bool { return s.State == cluster.Ready })
}

func TestAFailedBuildLeavesTheClusterFailed(t *testing.T) {
	p := newFakeProvisioner()
	c := cluster.New()

	NewCreateCluster(c, p, clock).Execute(context.Background())
	p.created <- errors.New("kind: node image pull failed")

	waitFor(t, c, func(s cluster.Status) bool { return s.State == cluster.Failed })
}

func TestCreateClusterIsBusyWhileDeleting(t *testing.T) {
	p := newFakeProvisioner()
	c := cluster.New()
	NewCreateCluster(c, p, clock).Execute(context.Background())
	p.created <- nil
	waitFor(t, c, func(s cluster.Status) bool { return s.State == cluster.Ready })
	NewDeleteCluster(c, p).Execute(context.Background())

	_, err := NewCreateCluster(c, p, clock).Execute(context.Background())

	if !errors.Is(err, cluster.ErrBusy) {
		t.Errorf("got %v, want ErrBusy", err)
	}
	p.deleted <- nil
}

var now = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)

func clock() time.Time { return now }

func waitFor(t *testing.T, c *cluster.Cluster, done func(cluster.Status) bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !done(c.Status()) {
		if time.Now().After(deadline) {
			t.Fatalf("still %+v after 5s", c.Status())
		}
		time.Sleep(time.Millisecond)
	}
}
