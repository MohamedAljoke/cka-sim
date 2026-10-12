package assistant

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

type Lab interface {
	Run(ctx context.Context, host, taskID string, script []byte) (string, error)
}

type Connection struct {
	MCPURL string
	Token  string
}

type Provider interface {
	ID() string
	Prepare(ctx context.Context, lab Lab, conn Connection) error
	Command() string
}

type Tokens interface {
	Issue(ctx context.Context, r *http.Request) (string, error)
}

func Find(providers []Provider, id string) (Provider, error) {
	for _, p := range providers {
		if p.ID() == id {
			return p, nil
		}
	}
	return nil, domainerr.Unprocessable("assistant_unknown", fmt.Sprintf("no assistant %q", id))
}
