package sandbox

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

var ErrNoLab = domainerr.Conflict("no_lab", "no lab is running; start one first")

type Provider interface {
	Name() string
	Create(ctx context.Context) (Box, error)
	Destroy(ctx context.Context, b Box) error
}

type Box struct {
	ID     string
	Runner tasks.Runner
	Shells labagent.Opener
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

type Lab struct {
	provider Provider
	now      func() time.Time
	mu       sync.Mutex
	status   Status
	box      Box
	settled  chan struct{}
}

func NewLab(p Provider, now func() time.Time) *Lab {
	return &Lab{provider: p, now: now, status: Status{State: None, Provider: p.Name()}}
}

func (l *Lab) Status() Status {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.status
}

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

func (l *Lab) Open(ctx context.Context, size labagent.Size) (labagent.Shell, error) {
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

func ClusterAnswers(ctx context.Context, b Box) error {
	_, err := b.Runner.Run(ctx, labagent.ControlPlaneNode, "health", []byte("kubectl get --raw=/readyz\n"))
	return err
}
