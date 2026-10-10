// Package sandbox gives the practice page and the exam a place to run: a box with a cluster,
// a way to run task scripts on it, and a way to open shells on it.
package sandbox

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/terminal"
)

var ErrNoLab = errors.New("no lab is running; start one first")

type Provider interface {
	Name() string
	// Create returns once the box's cluster is usable.
	Create(ctx context.Context) (Box, error)
	Destroy(ctx context.Context, b Box) error
}

type Box struct {
	ID     string
	Runner tasks.Runner
	Shells terminal.Opener
}

type State string

const (
	None     State = "none"
	Starting State = "starting"
	Ready    State = "ready"
	Failed   State = "failed"
)

type Status struct {
	State    State     `json:"state"`
	Provider string    `json:"provider"`
	Started  time.Time `json:"started,omitzero"`
	Ready    time.Time `json:"ready,omitzero"`
	Error    string    `json:"error,omitempty"`
}

// Lab holds at most one box. It is itself a tasks.Runner and a terminal.Opener, so the rest of
// the engine runs on whatever box is current without knowing where it lives.
type Lab struct {
	provider Provider
	now      func() time.Time
	mu       sync.Mutex
	status   Status
	box      Box
	// settled is closed when the current start succeeds or fails.
	settled chan struct{}
}

func NewLab(p Provider, now func() time.Time) *Lab {
	return &Lab{provider: p, now: now, status: Status{State: None, Provider: p.Name()}}
}

func (l *Lab) Status() Status {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.status
}

// Start creates a box in the background, unless one is ready or on its way.
func (l *Lab) Start() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.startLocked()
}

func (l *Lab) startLocked() chan struct{} {
	if l.status.State == Starting || l.status.State == Ready {
		return l.settled
	}
	settled := make(chan struct{})
	l.settled = settled
	l.status = Status{State: Starting, Provider: l.provider.Name(), Started: l.now()}
	go func() {
		// Not tied to a request: the page polls while the box is created.
		box, err := l.provider.Create(context.Background())
		l.mu.Lock()
		defer l.mu.Unlock()
		defer close(settled)
		if err != nil {
			l.status.State, l.status.Error = Failed, err.Error()
			log.Printf("lab on %s failed: %v", l.status.Provider, err)
			return
		}
		l.box = box
		l.status.State, l.status.Ready = Ready, l.now()
		log.Printf("lab on %s ready in %s", l.status.Provider, l.status.Ready.Sub(l.status.Started).Round(time.Millisecond))
	}()
	return settled
}

// Ensure starts the lab if needed and waits until it is ready.
func (l *Lab) Ensure(ctx context.Context) error {
	l.mu.Lock()
	settled := l.startLocked()
	l.mu.Unlock()
	select {
	case <-settled:
	case <-ctx.Done():
		return ctx.Err()
	}
	if s := l.Status(); s.State != Ready {
		return errors.New(s.Error)
	}
	return nil
}

func (l *Lab) End(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.status.State != Ready {
		return nil
	}
	if err := l.provider.Destroy(ctx, l.box); err != nil {
		return err
	}
	l.box = Box{}
	l.status = Status{State: None, Provider: l.provider.Name()}
	return nil
}

func (l *Lab) Run(ctx context.Context, host, taskID string, script []byte) (string, error) {
	box, err := l.current()
	if err != nil {
		return "", err
	}
	return box.Runner.Run(ctx, host, taskID, script)
}

func (l *Lab) Open(ctx context.Context, size terminal.Size) (terminal.Shell, error) {
	box, err := l.current()
	if err != nil {
		return nil, err
	}
	return box.Shells.Open(ctx, size)
}

func (l *Lab) EndAll(ctx context.Context) error {
	box, err := l.current()
	if err != nil {
		return nil
	}
	return box.Shells.EndAll(ctx)
}

func (l *Lab) current() (Box, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.status.State != Ready {
		return Box{}, ErrNoLab
	}
	return l.box, nil
}
