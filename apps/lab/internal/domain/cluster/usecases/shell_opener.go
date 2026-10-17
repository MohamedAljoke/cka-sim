package usecases

import (
	"context"
	"io"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

type ShellOpener interface {
	Open(ctx context.Context, size cluster.TerminalSize) (Shell, error)
}

type Shell interface {
	io.ReadWriter
	Resize(size cluster.TerminalSize) error
	Close() error
}
