package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type SolveTask struct {
	repo     TaskRepository
	runner   ScriptRunner
	clusters *clusters
}

func NewSolveTask(repo TaskRepository, runner ScriptRunner, clusters *clusters) *SolveTask {
	return &SolveTask{repo: repo, runner: runner, clusters: clusters}
}

// Execute runs the task's solution.sh on its host. A solution that fails is still an answer, as a setup is.
func (uc *SolveTask) Execute(ctx context.Context, owner, id string) (labs.ScriptResult, error) {
	task, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return labs.ScriptResult{}, err
	}

	if err != nil {
		return labs.ScriptResult{}, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	ctx, cancel := context.WithTimeout(ctx, setupTimeout)
	defer cancel()

	cluster := uc.clusters.of(owner)

	cluster.scripts.RLock()
	defer cluster.scripts.RUnlock()

	result, err := uc.runner.RunScript(ctx, owner, task.Host, task.Solve)
	switch {
	case errors.Is(err, labs.ErrNotReady):
		return labs.ScriptResult{}, err
	case err != nil:
		return labs.ScriptResult{}, fmt.Errorf("%w: solve %s: %w", labs.ErrUnavailable, task.ID, err)
	}

	return result, nil
}
