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
func (uc *GetCurrentTask) Execute(ctx context.Context, owner string) (tasks.Current, error) {
	id, setup := uc.clusters.of(owner).currentSetup()
	if id == "" {
		return tasks.Current{}, tasks.ErrNoCurrentTask
	}

	task, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return tasks.Current{}, err
	}

	if err != nil {
		return tasks.Current{}, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	return tasks.Current{Task: task, Setup: setup}, nil
}
