// Package terminal opens interactive shells on the cluster for the browser terminal.
package terminal

import (
	"context"
	"io"
)

type Size struct {
	Cols, Rows uint
}

type Opener interface {
	Open(ctx context.Context, size Size) (Shell, error)
	EndAll(ctx context.Context) error
}

type Shell interface {
	io.ReadWriter
	Resize(ctx context.Context, size Size) error
	Close() error
}
