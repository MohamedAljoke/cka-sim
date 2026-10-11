package local

import (
	"context"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

// Ending the lab keeps the container.
type Provider struct {
	Agent labagent.Client
}

func (Provider) Name() string { return "Local Docker" }

func (p Provider) Create(ctx context.Context) (sandbox.Box, error) {
	if err := p.Agent.Healthy(ctx); err != nil {
		return sandbox.Box{}, fmt.Errorf("no lab answers at %s; run make up: %w", p.Agent.URL, err)
	}
	box := sandbox.Box{ID: "local", Runner: p.Agent, Shells: p.Agent}
	if err := sandbox.ClusterAnswers(ctx, box); err != nil {
		return sandbox.Box{}, fmt.Errorf("the lab's cluster doesn't answer: %w", err)
	}
	// A server that crashed can leave its shells behind.
	if err := box.Shells.EndAll(ctx); err != nil {
		return sandbox.Box{}, err
	}
	return box, nil
}

func (Provider) Destroy(ctx context.Context, b sandbox.Box) error {
	return b.Shells.EndAll(ctx)
}
