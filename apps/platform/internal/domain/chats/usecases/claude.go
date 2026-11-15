package usecases

import (
	"context"
	"io"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

// Claude runs Claude Code in a terminal for a student, where the assistant runs, reaching only the tutor's tools.
type Claude interface {
	OpenClaude(ctx context.Context, student string, size labs.TerminalSize) (Terminal, error)
}

type Terminal interface {
	io.ReadWriter
	Resize(size labs.TerminalSize) error
	Close() error
}
