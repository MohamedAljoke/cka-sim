package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

const tidyTimeout = time.Minute

type TidyTask struct {
	repo    TaskRepository
	runner  ScriptRunner
	cluster *cluster
}

func NewTidyTask(repo TaskRepository, runner ScriptRunner, cluster *cluster) *TidyTask {
	return &TidyTask{repo: repo, runner: runner, cluster: cluster}
}

// Execute undoes what each task broke, then deletes the namespaces its setup made without waiting for them to go.
// The tasks are tidied together, so no setup runs between two of them. A reset runs even on a task that isn't
// touched: a restarted server no longer knows what was set up.
func (uc *TidyTask) Execute(ctx context.Context, ids ...string) error {
	all := make([]tasks.Task, 0, len(ids))

	for _, id := range ids {
		task, err := uc.repo.Get(ctx, id)
		if errors.Is(err, tasks.ErrNotFound) {
			return err
		}

		if err != nil {
			return fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
		}

		all = append(all, task)
	}

	uc.cluster.scripts.Lock()
	defer uc.cluster.scripts.Unlock()

	var errs []error

	for _, task := range all {
		err := uc.tidy(ctx, task)
		if errors.Is(err, labs.ErrNotReady) || errors.Is(err, labs.ErrUnavailable) {
			return err
		}

		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (uc *TidyTask) tidy(ctx context.Context, task tasks.Task) error {
	if task.Reset != "" {
		result, err := reset(ctx, uc.runner, task)
		if err != nil {
			return err
		}

		if !result.Passed() {
			uc.cluster.touch(task.ID)
			return fmt.Errorf("reset %s exited with code %d:\n%s", task.ID, result.ExitCode, result.Output)
		}
	}

	uc.cluster.forget(task.ID)

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), tidyTimeout)
	defer cancel()

	result, err := uc.runner.RunScript(ctx, task.Host, TidyScript(task.ID))
	switch {
	case errors.Is(err, labs.ErrNotReady):
		return err
	case err != nil:
		return fmt.Errorf("%w: tidy %s: %w", labs.ErrUnavailable, task.ID, err)
	case !result.Passed():
		return fmt.Errorf("delete the namespaces of %s: exit code %d:\n%s", task.ID, result.ExitCode, result.Output)
	}

	return nil
}

// TidyScript deletes the namespaces fresh_ns labelled with the task's id.
func TidyScript(id string) string {
	return "kubectl delete namespace -l cka-sim/task=" + id + " --wait=false >/dev/null\n"
}
