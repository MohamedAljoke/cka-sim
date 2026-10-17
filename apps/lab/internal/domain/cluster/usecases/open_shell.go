package usecases

import (
	"context"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

type OpenShell struct {
	cluster *cluster.Cluster
	opener  ShellOpener
}

func NewOpenShell(c *cluster.Cluster, o ShellOpener) *OpenShell {
	return &OpenShell{cluster: c, opener: o}
}

func (uc *OpenShell) Execute(ctx context.Context, size cluster.TerminalSize) (Shell, error) {
	if uc.cluster.Status().State != cluster.Ready {
		return nil, cluster.ErrNotReady
	}

	shell, err := uc.opener.Open(ctx, size)
	if err != nil {
		return nil, fmt.Errorf("open a shell: %w", err)
	}

	return shell, nil
}
