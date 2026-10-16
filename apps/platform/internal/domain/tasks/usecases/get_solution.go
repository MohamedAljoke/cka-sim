package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type GetSolution struct {
	repo TaskRepository
}

func NewGetSolution(repo TaskRepository) *GetSolution {
	return &GetSolution{repo: repo}
}

func (uc *GetSolution) Execute(ctx context.Context, id string) (string, error) {
	t, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}
	return t.Solution, nil
}
