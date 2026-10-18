package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type OpenShell struct {
	provider LabProvider
}

func NewOpenShell(provider LabProvider) *OpenShell {
	return &OpenShell{provider: provider}
}

func (uc *OpenShell) Execute(ctx context.Context, owner string, size labs.TerminalSize) (Shell, error) {
	shell, err := uc.provider.OpenShell(ctx, owner, size)
	switch {
	case errors.Is(err, labs.ErrNotReady):
		return nil, err
	case err != nil:
		return nil, fmt.Errorf("%w: %w", labs.ErrUnavailable, err)
	}

	return shell, nil
}
