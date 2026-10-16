package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type GetTask struct {
	repo TaskRepository
}

func NewGetTask(repo TaskRepository) *GetTask {
	return &GetTask{repo: repo}
}

func (uc *GetTask) Execute(ctx context.Context, id string) (tasks.Task, error) {
	task, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return tasks.Task{}, err
	}

	if err != nil {
		return tasks.Task{}, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	return task, nil
}
