// Package fly hands out exam boxes from a pool of suspended Fly Machines, each with its cluster
// already up. Filling the pool is infrastructure work (deploy/fly/pool.sh), not this package's.
package fly

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
)

var ErrPoolEmpty = errors.New("no warm VM in the pool; fill it with deploy/fly/pool.sh fill")

const (
	poolKey = "pool"
	ready   = "ready"
	claimed = "claimed"
	// Each pool VM's dockerd listens on its private address only.
	dockerPort    = "2375"
	createTimeout = 3 * time.Minute
)

type Config struct {
	// App holds the pool's Machines.
	App   string
	Token string
	Lib   []byte
	// API defaults to the public Machines API.
	API string
}

type Provider struct {
	api        machines
	healthWait time.Duration
	box        func(ctx context.Context, id, host string) (sandbox.Box, error)
	probe      func(ctx context.Context, b sandbox.Box) error
}

func New(c Config) *Provider {
	base := c.API
	if base == "" {
		base = "https://api.machines.dev"
	}
	return &Provider{
		api:        machines{base: base + "/v1/apps/" + c.App, token: c.Token, client: http.DefaultClient},
		healthWait: 30 * time.Second,
		box: func(ctx context.Context, id, host string) (sandbox.Box, error) {
			return sandbox.DockerBox(ctx, id, host, c.Lib)
		},
		probe: clusterAnswers,
	}
}

func (*Provider) Name() string { return "Fly" }

// Create claims a warm Machine and resumes it. One whose cluster doesn't answer, say because Fly
// cold-booted it instead of restoring its snapshot, is destroyed and the next one is tried.
func (p *Provider) Create(ctx context.Context) (sandbox.Box, error) {
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()
	warm, err := p.api.list(ctx, poolKey, ready)
	if err != nil {
		return sandbox.Box{}, err
	}
	if len(warm) == 0 {
		return sandbox.Box{}, ErrPoolEmpty
	}
	for _, m := range warm {
		box, err := p.claim(ctx, m)
		if err == nil {
			return box, nil
		}
		log.Printf("fly: machine %s is unusable, destroying it: %v", m.ID, err)
		if err := p.api.destroy(context.WithoutCancel(ctx), m.ID); err != nil {
			log.Printf("fly: destroy machine %s: %v", m.ID, err)
		}
	}
	return sandbox.Box{}, fmt.Errorf("none of the %d warm VMs came up healthy", len(warm))
}

func (p *Provider) claim(ctx context.Context, m flyMachine) (sandbox.Box, error) {
	if err := p.api.setMetadata(ctx, m.ID, poolKey, claimed); err != nil {
		return sandbox.Box{}, err
	}
	if err := p.api.start(ctx, m.ID); err != nil {
		return sandbox.Box{}, err
	}
	if err := p.api.waitStarted(ctx, m.ID); err != nil {
		return sandbox.Box{}, err
	}
	return p.healthy(ctx, m.ID, "tcp://["+m.PrivateIP+"]:"+dockerPort)
}

// healthy retries for a while: dockerd and the API server can take a moment after a resume.
func (p *Provider) healthy(ctx context.Context, id, host string) (sandbox.Box, error) {
	deadline := time.Now().Add(p.healthWait)
	for {
		box, err := p.box(ctx, id, host)
		if err == nil {
			if err = p.probe(ctx, box); err == nil {
				return box, nil
			}
		}
		if time.Now().After(deadline) {
			return sandbox.Box{}, err
		}
		select {
		case <-ctx.Done():
			return sandbox.Box{}, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func (p *Provider) Destroy(ctx context.Context, b sandbox.Box) error {
	return p.api.destroy(ctx, b.ID)
}

// ReapClaimed destroys Machines a previous run of the server claimed and never gave back.
func (p *Provider) ReapClaimed(ctx context.Context) error {
	left, err := p.api.list(ctx, poolKey, claimed)
	if err != nil {
		return err
	}
	for _, m := range left {
		if err := p.api.destroy(ctx, m.ID); err != nil {
			return err
		}
		log.Printf("fly: destroyed machine %s, left claimed by an earlier run", m.ID)
	}
	return nil
}

func clusterAnswers(ctx context.Context, b sandbox.Box) error {
	_, err := b.Runner.Run(ctx, cluster.ControlPlaneNode, "health", []byte("kubectl get --raw=/readyz\n"))
	return err
}
