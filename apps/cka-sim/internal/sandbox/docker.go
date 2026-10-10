package sandbox

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/runner"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/terminal"
)

// DockerBox is a box whose cluster runs on the Docker daemon at host, on this machine or on a VM;
// an empty host means docker's own default.
func DockerBox(ctx context.Context, id, host string, lib []byte) (Box, error) {
	shells, err := terminal.NewDockerOpener(host, cluster.Base, cluster.Candidate)
	if err != nil {
		return Box{}, err
	}
	// A server that crashed can leave its shells behind.
	if err := shells.EndAll(ctx); err != nil {
		return Box{}, err
	}
	return Box{ID: id, Runner: runner.Node{Lib: lib, DockerHost: host}, Shells: shells}, nil
}
