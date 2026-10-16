package usecases

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type ListTasks struct {
	repo TaskRepository
}

func NewListTasks(repo TaskRepository) *ListTasks {
	return &ListTasks{repo: repo}
}

func (uc *ListTasks) Execute(ctx context.Context) ([]tasks.Task, error) {
	all, err := uc.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}
	sorted := slices.Clone(all)
	slices.SortFunc(sorted, func(a, b tasks.Task) int { return strings.Compare(a.ID, b.ID) })
	return sorted, nil
}
