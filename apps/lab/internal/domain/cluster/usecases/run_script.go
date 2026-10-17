package usecases

import (
	"context"
	"fmt"
	"slices"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

type RunScript struct {
	cluster     *cluster.Cluster
	provisioner Provisioner
	runner      ScriptRunner
}

func NewRunScript(c *cluster.Cluster, p Provisioner, r ScriptRunner) *RunScript {
	return &RunScript{cluster: c, provisioner: p, runner: r}
}

func (uc *RunScript) Execute(ctx context.Context, host, script string) (cluster.ScriptResult, error) {
	if uc.cluster.Status().State != cluster.Ready {
		return cluster.ScriptResult{}, cluster.ErrNotReady
	}

	nodes, err := uc.provisioner.Nodes(ctx)
	if err != nil {
		return cluster.ScriptResult{}, fmt.Errorf("list the cluster's nodes: %w", err)
	}

	if !slices.Contains(nodes, host) {
		return cluster.ScriptResult{}, cluster.ErrUnknownHost
	}

	result, err := uc.runner.Run(ctx, host, script)
	if err != nil {
		return cluster.ScriptResult{}, fmt.Errorf("run a script on %s: %w", host, err)
	}

	return result, nil
}
