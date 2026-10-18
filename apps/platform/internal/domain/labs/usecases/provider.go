package usecases

import (
	"context"
	"io"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

// LabProvider holds a lab per owner: each student's own cluster.
type LabProvider interface {
	Get(ctx context.Context, owner string) (labs.Lab, error)
	Start(ctx context.Context, owner string) (labs.Lab, error)
	Stop(ctx context.Context, owner string) (labs.Lab, error)
	OpenShell(ctx context.Context, owner string, size labs.TerminalSize) (Shell, error)
}

type Shell interface {
	io.ReadWriter
	Resize(size labs.TerminalSize) error
	Close() error
}
