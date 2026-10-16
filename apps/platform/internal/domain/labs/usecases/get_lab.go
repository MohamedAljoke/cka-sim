package usecases

import (
	"context"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type GetLab struct {
	provider LabProvider
}

func NewGetLab(provider LabProvider) *GetLab {
	return &GetLab{provider: provider}
}

func (uc *GetLab) Execute(ctx context.Context) (labs.Lab, error) {
	lab, err := uc.provider.Get(ctx)
	if err != nil {
		return labs.Lab{}, fmt.Errorf("%w: %w", labs.ErrUnavailable, err)
	}
	return lab, nil
}
