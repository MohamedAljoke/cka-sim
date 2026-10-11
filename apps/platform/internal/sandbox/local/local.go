package local

import (
	"context"
	"fmt"
	"sync"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

// One container is one cluster, so only one account can hold it: a second one would share its
// namespaces, and the shell cleanup in Create and Destroy would close the first one's terminal.
var ErrInUse = domainerr.Conflict("lab_in_use", "another account is using the local lab; it runs one at a time")

// Ending the lab keeps the container.
type Provider struct {
	Agent labagent.Client
	mu    sync.Mutex
	owner string
}

func (*Provider) Name() string { return "Local Docker" }

func (p *Provider) Create(ctx context.Context, owner string) (sandbox.Box, error) {
	if err := p.claim(owner); err != nil {
		return sandbox.Box{}, err
	}
	box, err := p.reach(ctx, owner)
	if err != nil {
		p.release(owner)
		return sandbox.Box{}, err
	}
	return box, nil
}

func (p *Provider) Destroy(ctx context.Context, b sandbox.Box) error {
	if err := b.Shells.EndAll(ctx); err != nil {
		return err
	}
	p.release(b.ID)
	return nil
}

func (p *Provider) reach(ctx context.Context, owner string) (sandbox.Box, error) {
	if err := p.Agent.Healthy(ctx); err != nil {
		return sandbox.Box{}, fmt.Errorf("no lab answers at %s; run make up: %w", p.Agent.URL, err)
	}
	box := sandbox.Box{ID: owner, Runner: p.Agent, Shells: p.Agent}
	if err := sandbox.ClusterAnswers(ctx, box); err != nil {
		return sandbox.Box{}, fmt.Errorf("the lab's cluster doesn't answer: %w", err)
	}
	// A server that crashed can leave its shells behind.
	if err := box.Shells.EndAll(ctx); err != nil {
		return sandbox.Box{}, err
	}
	return box, nil
}

func (p *Provider) claim(owner string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.owner != "" && p.owner != owner {
		return ErrInUse
	}
	p.owner = owner
	return nil
}

func (p *Provider) release(owner string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.owner == owner {
		p.owner = ""
	}
}
