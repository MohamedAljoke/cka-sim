package usecases

import (
	"context"
	"io"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude"
)

// CLI starts Claude Code in a terminal that reaches no tools but the MCP endpoint it is given.
type CLI interface {
	Start(ctx context.Context, launch Launch) (Session, error)
}

type Launch struct {
	Student string
	MCP     Endpoint
	Size    claude.TerminalSize
}

type Session interface {
	io.ReadWriter
	Resize(size claude.TerminalSize) error
	Close() error
}

// Endpoints serve the tutor's MCP tools for one student at a URL that only the holder of its token reaches.
type Endpoints interface {
	Open(student string) Endpoint
	Close(endpoint Endpoint)
}

type Endpoint struct {
	URL   string
	Token string
}
