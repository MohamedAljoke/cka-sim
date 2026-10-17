package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

func TestAScriptNeedsAReadyCluster(t *testing.T) {
	r := &fakeScriptRunner{}

	_, err := NewRunScript(cluster.New(), newFakeProvisioner(), r).Execute(context.Background(), "lab-control-plane", "true")

	if !errors.Is(err, cluster.ErrNotReady) || r.ran {
		t.Errorf("got %v, ran %v", err, r.ran)
	}
}

func TestAScriptRunsOnlyOnTheClustersNodes(t *testing.T) {
	p := newFakeProvisioner()
	c := readyCluster(t, p)
	p.nodes = []string{"lab-control-plane", "lab-worker"}
	r := &fakeScriptRunner{}

	_, err := NewRunScript(c, p, r).Execute(context.Background(), "my-database", "true")

	if !errors.Is(err, cluster.ErrUnknownHost) || r.ran {
		t.Errorf("got %v, ran %v", err, r.ran)
	}
}

func TestAScriptAnswersWhatTheRunnerSays(t *testing.T) {
	p := newFakeProvisioner()
	c := readyCluster(t, p)
	p.nodes = []string{"lab-control-plane", "lab-worker"}
	r := &fakeScriptRunner{result: cluster.ScriptResult{Output: []byte("no such file\n"), ExitCode: 1}}

	got, err := NewRunScript(c, p, r).Execute(context.Background(), "lab-worker", "cat /nope")

	if err != nil || string(got.Output) != "no such file\n" || got.ExitCode != 1 {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestAScriptKeepsTheCauseWhenTheNodesCannotBeListed(t *testing.T) {
	p := newFakeProvisioner()
	c := readyCluster(t, p)
	p.nodesErr = errors.New("docker: not found")

	_, err := NewRunScript(c, p, &fakeScriptRunner{}).Execute(context.Background(), "lab-worker", "true")

	if !errors.Is(err, p.nodesErr) {
		t.Errorf("got %v", err)
	}
}
