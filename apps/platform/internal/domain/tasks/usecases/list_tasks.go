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

func (uc *ListTasks) Execute(ctx context.Context, filter tasks.Filter) ([]tasks.Task, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	all, err := uc.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}
	kept := []tasks.Task{}
	for _, t := range all {
		if filter.Matches(t) {
			kept = append(kept, t)
		}
	}
	slices.SortFunc(kept, func(a, b tasks.Task) int { return strings.Compare(a.ID, b.ID) })
	return kept, nil
}
