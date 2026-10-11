package gke

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

const (
	agentPort = "7171"
	// A warm pool answers in a second; an empty one waits for a new node and its cluster.
	createTimeout = 15 * time.Minute
)

type Config struct {
	Pool       string
	AgentToken string
}

type Provider struct {
	kube       kube
	pool       string
	poll       time.Duration
	healthWait time.Duration
	box        func(id, ip string) sandbox.Box
	probe      func(ctx context.Context, b sandbox.Box) error
}

func New(c Config) (*Provider, error) {
	k, err := inCluster()
	if err != nil {
		return nil, err
	}
	return &Provider{
		kube:       k,
		pool:       c.Pool,
		poll:       time.Second,
		healthWait: 30 * time.Second,
		box: func(id, ip string) sandbox.Box {
			agent := labagent.Client{URL: "http://" + net.JoinHostPort(ip, agentPort), Token: c.AgentToken}
			return sandbox.Box{ID: id, Runner: agent, Shells: agent}
		},
		probe: sandbox.ClusterAnswers,
	}, nil
}

func (*Provider) Name() string { return "GKE" }

// A claim that never becomes usable is deleted, so it doesn't hold a node.
func (p *Provider) Create(ctx context.Context, owner string) (sandbox.Box, error) {
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()
	name, err := p.kube.createClaim(ctx, p.pool, owner)
	if err != nil {
		return sandbox.Box{}, err
	}
	box, err := p.claimed(ctx, name)
	if err != nil {
		if err := p.kube.deleteClaim(context.WithoutCancel(ctx), name); err != nil {
			log.Printf("gke: delete claim %s: %v", name, err)
		}
		return sandbox.Box{}, err
	}
	return box, nil
}

func (p *Provider) claimed(ctx context.Context, name string) (sandbox.Box, error) {
	sandboxName, err := p.waitReady(ctx, name)
	if err != nil {
		return sandbox.Box{}, err
	}
	// The sandbox's pod has the sandbox's name.
	ip, err := p.kube.podIP(ctx, sandboxName)
	if err != nil {
		return sandbox.Box{}, err
	}
	box := p.box(name, ip)
	if err := p.healthy(ctx, box); err != nil {
		return sandbox.Box{}, fmt.Errorf("lab %s doesn't answer: %w", sandboxName, err)
	}
	return box, nil
}

func (p *Provider) waitReady(ctx context.Context, name string) (string, error) {
	for {
		c, err := p.kube.getClaim(ctx, name)
		if ctx.Err() != nil {
			return "", fmt.Errorf("claim %s never became ready: %w", name, ctx.Err())
		}
		if err != nil {
			return "", err
		}
		if c.ready() {
			return c.Status.Sandbox.Name, nil
		}
		select {
		case <-ctx.Done():
		case <-time.After(p.poll):
		}
	}
}

// healthy retries for a while: a pod can be Ready a moment before its network policy lets us in.
func (p *Provider) healthy(ctx context.Context, box sandbox.Box) error {
	deadline := time.Now().Add(p.healthWait)
	for {
		err := p.probe(ctx, box)
		if err == nil || time.Now().After(deadline) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(p.poll):
		}
	}
}

func (p *Provider) Destroy(ctx context.Context, b sandbox.Box) error {
	return p.kube.deleteClaim(ctx, b.ID)
}

func (p *Provider) ReapClaims(ctx context.Context) error {
	names, err := p.kube.ourClaims(ctx)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := p.kube.deleteClaim(ctx, name); err != nil {
			return err
		}
		log.Printf("gke: deleted claim %s, left by an earlier run", name)
	}
	return nil
}
