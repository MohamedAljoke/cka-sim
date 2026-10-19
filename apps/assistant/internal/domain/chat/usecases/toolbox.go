package usecases

import (
	"context"
	"encoding/json"
)

// Toolbox opens the tutor's tools for one student.
type Toolbox interface {
	Open(ctx context.Context, student string) (Tools, error)
}

type Tools interface {
	// Instructions say who the tutor is and how it uses the tools.
	Instructions() string
	List() []Tool
	// Call answers a tool's output; failed means the tool refused, and the output says why.
	Call(ctx context.Context, name string, args json.RawMessage) (output string, failed bool, err error)
	Close() error
}
