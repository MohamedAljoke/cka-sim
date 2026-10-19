package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type GetCurrentTask struct {
	repo     TaskRepository
	clusters *clusters
}

func NewGetCurrentTask(repo TaskRepository, clusters *clusters) *GetCurrentTask {
	return &GetCurrentTask{repo: repo, clusters: clusters}
}

// Execute answers the task the owner set up last and hasn't left. A restarted server has forgotten it.
func (uc *GetCurrentTask) Execute(ctx context.Context, owner string) (tasks.Task, error) {
	id := uc.clusters.of(owner).currentTask()
	if id == "" {
		return tasks.Task{}, tasks.ErrNoCurrentTask
	}

	task, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return tasks.Task{}, err
	}

	if err != nil {
		return tasks.Task{}, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	return task, nil
}
