package labagent

import (
	"context"
	"sync"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
)

// Fleet holds each owner's lab, on the machine machineOf gives that owner: one pod per student on Kubernetes, or
// the same container for everyone on one person's machine.
type Fleet struct {
	machineOf func(owner string) Machine
	token     string

	mu   sync.Mutex
	labs map[string]*Lab
}

func NewFleet(machineOf func(owner string) Machine, token string) *Fleet {
	return &Fleet{machineOf: machineOf, token: token, labs: map[string]*Lab{}}
}

func (f *Fleet) Get(ctx context.Context, owner string) (labs.Lab, error) {
	return f.lab(owner).Get(ctx)
}

func (f *Fleet) Start(ctx context.Context, owner string) (labs.Lab, error) {
	return f.lab(owner).Start(ctx)
}

func (f *Fleet) Stop(ctx context.Context, owner string) (labs.Lab, error) {
	return f.lab(owner).Stop(ctx)
}

func (f *Fleet) OpenShell(ctx context.Context, owner string, size labs.TerminalSize) (usecases.Shell, error) {
	return f.lab(owner).OpenShell(ctx, size)
}

func (f *Fleet) RunScript(ctx context.Context, owner, host, script string) (labs.ScriptResult, error) {
	return f.lab(owner).RunScript(ctx, host, script)
}

func (f *Fleet) lab(owner string) *Lab {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.labs[owner] == nil {
		f.labs[owner] = NewLab(f.machineOf(owner), f.token)
	}

	return f.labs[owner]
}
