package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

const resetTimeout = 3 * time.Minute

type HealCluster struct {
	repo    TaskRepository
	runner  ScriptRunner
	cluster *cluster
}

func NewHealCluster(repo TaskRepository, runner ScriptRunner, cluster *cluster) *HealCluster {
	return &HealCluster{repo: repo, runner: runner, cluster: cluster}
}

// Execute resets every task set up since its reset last ran, one at a time, and answers with the first reset that
// failed. A task whose reset failed stays touched, so the next heal tries it again.
func (uc *HealCluster) Execute(ctx context.Context) (labs.ScriptResult, error) {
	all, err := uc.repo.List(ctx)
	if err != nil {
		return labs.ScriptResult{}, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	uc.cluster.scripts.Lock()
	defer uc.cluster.scripts.Unlock()

	var failed labs.ScriptResult

	for _, task := range all {
		if !uc.cluster.isTouched(task.ID) {
			continue
		}

		if task.Reset == "" {
			uc.cluster.forget(task.ID)
			continue
		}

		result, err := reset(ctx, uc.runner, task)
		if err != nil {
			return labs.ScriptResult{}, err
		}

		if !result.Passed() {
			if failed.Passed() {
				failed = labs.ScriptResult{ExitCode: result.ExitCode,
					Output: fmt.Sprintf("%s was set up before, and its reset.sh failed:\n%s", task.ID, result.Output)}
			}

			continue
		}

		uc.cluster.forget(task.ID)
	}

	return failed, nil
}

// A reset left half done leaves the cluster broken, so it runs to the end even when the page that asked goes away.
func reset(ctx context.Context, runner ScriptRunner, task tasks.Task) (labs.ScriptResult, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), resetTimeout)
	defer cancel()

	result, err := runner.RunScript(ctx, task.Host, task.Reset)
	switch {
	case errors.Is(err, labs.ErrNotReady):
		return labs.ScriptResult{}, err
	case err != nil:
		return labs.ScriptResult{}, fmt.Errorf("%w: reset %s: %w", labs.ErrUnavailable, task.ID, err)
	}

	return result, nil
}
