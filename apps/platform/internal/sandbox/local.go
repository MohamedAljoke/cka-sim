package sandbox

import (
	"context"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/lab/api"
)

// Local is the lab container that make up runs on this machine, reached through its agent like a lab on
// GKE. Ending the lab keeps the container.
type Local struct {
	Agent api.Client
}

func (Local) Name() string { return "Local Docker" }

func (l Local) Create(ctx context.Context) (Box, error) {
	if err := l.Agent.Healthy(ctx); err != nil {
		return Box{}, fmt.Errorf("no lab answers at %s; run make up: %w", l.Agent.URL, err)
	}
	box := Box{ID: "local", Runner: l.Agent, Shells: l.Agent}
	if err := ClusterAnswers(ctx, box); err != nil {
		return Box{}, fmt.Errorf("the lab's cluster doesn't answer: %w", err)
	}
	// A server that crashed can leave its shells behind.
	if err := box.Shells.EndAll(ctx); err != nil {
		return Box{}, err
	}
	return box, nil
}

func (Local) Destroy(ctx context.Context, b Box) error {
	return b.Shells.EndAll(ctx)
}

// ClusterAnswers proves a box is usable: kubectl on the control plane, through the box's runner.
func ClusterAnswers(ctx context.Context, b Box) error {
	_, err := b.Runner.Run(ctx, api.ControlPlaneNode, "health", []byte("kubectl get --raw=/readyz\n"))
	return err
}
