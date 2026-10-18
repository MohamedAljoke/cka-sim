package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type StartLab struct {
	provider LabProvider
}

func NewStartLab(provider LabProvider) *StartLab {
	return &StartLab{provider: provider}
}

func (uc *StartLab) Execute(ctx context.Context, owner string) (labs.Lab, error) {
	lab, err := uc.provider.Start(ctx, owner)
	switch {
	case errors.Is(err, labs.ErrBusy):
		return labs.Lab{}, err
	case err != nil:
		return labs.Lab{}, fmt.Errorf("%w: %w", labs.ErrUnavailable, err)
	}

	return lab, nil
}
