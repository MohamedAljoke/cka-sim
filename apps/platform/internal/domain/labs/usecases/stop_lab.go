package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type StopLab struct {
	provider LabProvider
}

func NewStopLab(provider LabProvider) *StopLab {
	return &StopLab{provider: provider}
}

func (uc *StopLab) Execute(ctx context.Context, owner string) (labs.Lab, error) {
	lab, err := uc.provider.Stop(ctx, owner)
	switch {
	case errors.Is(err, labs.ErrBusy):
		return labs.Lab{}, err
	case err != nil:
		return labs.Lab{}, fmt.Errorf("%w: %w", labs.ErrUnavailable, err)
	}

	return lab, nil
}
