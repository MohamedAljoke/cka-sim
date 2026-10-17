package usecases

import (
	"context"
	"io"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type LabProvider interface {
	Get(ctx context.Context) (labs.Lab, error)
	Start(ctx context.Context) (labs.Lab, error)
	Stop(ctx context.Context) (labs.Lab, error)
	OpenShell(ctx context.Context, size labs.TerminalSize) (Shell, error)
}

type Shell interface {
	io.ReadWriter
	Resize(size labs.TerminalSize) error
	Close() error
}
