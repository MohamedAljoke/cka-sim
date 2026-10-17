package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

func TestAShellNeedsAReadyCluster(t *testing.T) {
	o := &fakeShellOpener{}
	_, err := NewOpenShell(cluster.New(), o).Execute(context.Background(), cluster.TerminalSize{Cols: 80, Rows: 24})

	if !errors.Is(err, cluster.ErrNotReady) {
		t.Errorf("got %v, want ErrNotReady", err)
	}
	if o.opened != nil {
		t.Errorf("opened a shell at %v on a cluster that isn't ready", *o.opened)
	}
}

func TestAShellOpensAtTheTerminalsSize(t *testing.T) {
	o := &fakeShellOpener{}
	size := cluster.TerminalSize{Cols: 120, Rows: 40}

	shell, err := NewOpenShell(readyCluster(t, newFakeProvisioner()), o).Execute(context.Background(), size)
	if err != nil || shell == nil {
		t.Fatalf("Execute = %v, %v", shell, err)
	}
	if o.opened == nil || *o.opened != size {
		t.Errorf("opened at %v, want %v", o.opened, size)
	}
}

func TestAShellThatCannotOpenKeepsTheCause(t *testing.T) {
	cause := errors.New("no such container")
	o := &fakeShellOpener{err: cause}

	_, err := NewOpenShell(readyCluster(t, newFakeProvisioner()), o).Execute(context.Background(), cluster.TerminalSize{Cols: 80, Rows: 24})

	if !errors.Is(err, cause) {
		t.Errorf("got %v, want it to wrap %v", err, cause)
	}
}
