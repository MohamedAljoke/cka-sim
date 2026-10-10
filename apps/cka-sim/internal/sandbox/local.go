package sandbox

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/cluster"
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
	return DockerBox(ctx, "local", "", l.Lib)
}

func (Local) Destroy(ctx context.Context, b Box) error {
	return b.Shells.EndAll(ctx)
}
