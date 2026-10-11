package assistant

import (
	"context"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

type Lab interface {
	Run(ctx context.Context, host, taskID string, script []byte) (string, error)
}

type Provider interface {
	ID() string
	Prepare(ctx context.Context, lab Lab) error
	Command() string
}

func Find(providers []Provider, id string) (Provider, error) {
	for _, p := range providers {
		if p.ID() == id {
			return p, nil
		}
	}
	return nil, domainerr.Unprocessable("assistant_unknown", fmt.Sprintf("no assistant %q", id))
}
