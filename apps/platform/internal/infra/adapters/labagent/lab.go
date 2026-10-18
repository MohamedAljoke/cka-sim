package labagent

import (
	"context"
	"fmt"
	"sync"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
)

// Machine is where a lab's agent runs: a container on this machine, or a pod made on demand.
type Machine interface {
	// Up makes sure the machine exists; it may take minutes before its agent answers.
	Up(ctx context.Context) error
	// Down removes the machine, cluster and all, and says whether it did: one that always exists stays.
	Down(ctx context.Context) (removed bool, err error)
	Where(ctx context.Context) (Place, error)
}

// Place is where a machine's agent answers. With no URL the agent can't answer yet, and Lab says why.
type Place struct {
	URL string
	Lab labs.Lab
	// The machine exists only because Up was asked for, so an agent with no cluster is told to build one.
	OnDemand bool
}

type Lab struct {
	machine Machine
	token   string

	mu    sync.Mutex
	url   string
	agent *agent
}

func NewLab(machine Machine, token string) *Lab {
	return &Lab{machine: machine, token: token}
}

func (l *Lab) Get(ctx context.Context) (labs.Lab, error) {
	place, err := l.machine.Where(ctx)
	if err != nil || place.URL == "" {
		return place.Lab, err
	}

	agent := l.agentAt(place.URL)

	lab, err := agent.Get(ctx)
	if err != nil || lab.State != labs.None || !place.OnDemand {
		return lab, err
	}

	return agent.Start(ctx)
}

func (l *Lab) Start(ctx context.Context) (labs.Lab, error) {
	if err := l.machine.Up(ctx); err != nil {
		return labs.Lab{}, err
	}

	place, err := l.machine.Where(ctx)

	switch {
	case err != nil:
		return labs.Lab{}, err
	case place.Lab.State == labs.Stopping:
		return labs.Lab{}, fmt.Errorf("%w: the last lab is still stopping", labs.ErrBusy)
	case place.URL == "":
		return place.Lab, nil
	}

	return l.agentAt(place.URL).Start(ctx)
}

func (l *Lab) Stop(ctx context.Context) (labs.Lab, error) {
	removed, err := l.machine.Down(ctx)
	switch {
	case err != nil:
		return labs.Lab{}, err
	case removed:
		return labs.Lab{State: labs.Stopping}, nil
	}

	place, err := l.machine.Where(ctx)
	if err != nil || place.URL == "" {
		return place.Lab, err
	}

	return l.agentAt(place.URL).Stop(ctx)
}

func (l *Lab) RunScript(ctx context.Context, host, script string) (labs.ScriptResult, error) {
	agent, err := l.reachable(ctx)
	if err != nil {
		return labs.ScriptResult{}, err
	}

	return agent.RunScript(ctx, host, script)
}

func (l *Lab) OpenShell(ctx context.Context, size labs.TerminalSize) (usecases.Shell, error) {
	agent, err := l.reachable(ctx)
	if err != nil {
		return nil, err
	}

	return agent.OpenShell(ctx, size)
}

func (l *Lab) reachable(ctx context.Context) (*agent, error) {
	place, err := l.machine.Where(ctx)
	switch {
	case err != nil:
		return nil, err
	case place.URL == "":
		return nil, fmt.Errorf("%w: the lab is %s", labs.ErrNotReady, place.Lab.State)
	}

	return l.agentAt(place.URL), nil
}

// A new machine can have a new address, so the client is made again only then.
func (l *Lab) agentAt(url string) *agent {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.agent == nil || l.url != url {
		l.url, l.agent = url, newAgent(url, l.token)
	}

	return l.agent
}

// Fixed is a machine that is always there, at one address: the lab container on this machine.
type Fixed string

func (Fixed) Up(context.Context) error { return nil }

func (Fixed) Down(context.Context) (bool, error) { return false, nil }

func (f Fixed) Where(context.Context) (Place, error) { return Place{URL: string(f)}, nil }
