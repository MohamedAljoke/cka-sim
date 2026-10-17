package usecases

import (
	"bytes"
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

type fakeShellOpener struct {
	opened *cluster.TerminalSize
	err    error
}

func (o *fakeShellOpener) Open(_ context.Context, size cluster.TerminalSize) (Shell, error) {
	if o.err != nil {
		return nil, o.err
	}
	o.opened = &size
	return &fakeShell{}, nil
}

type fakeShell struct {
	bytes.Buffer
}

func (s *fakeShell) Resize(cluster.TerminalSize) error { return nil }

func (s *fakeShell) Close() error { return nil }
