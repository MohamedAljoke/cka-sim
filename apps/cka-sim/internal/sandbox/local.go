package sandbox

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/runner"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/terminal"
)

// Local is the cluster that cka-sim up made in Docker on this machine. Ending the lab keeps it.
type Local struct {
	Lib []byte
}

func (Local) Name() string { return "Local Docker" }

func (l Local) Create(ctx context.Context) (Box, error) {
	if err := cluster.RequireUp(); err != nil {
		return Box{}, err
	}
	shells, err := terminal.NewDockerOpener(cluster.Base, cluster.Candidate)
	if err != nil {
		return Box{}, err
	}
	// A server that crashed can leave its shells behind.
	if err := shells.EndAll(ctx); err != nil {
		return Box{}, err
	}
	return Box{ID: "local", Runner: runner.Node{Lib: l.Lib}, Shells: shells}, nil
}

func (Local) Destroy(ctx context.Context, b Box) error {
	return b.Shells.EndAll(ctx)
}
