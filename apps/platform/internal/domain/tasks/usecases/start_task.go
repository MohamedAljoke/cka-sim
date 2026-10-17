package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

const setupTimeout = 3 * time.Minute

type StartTask struct {
	repo   TaskRepository
	runner ScriptRunner
}

func NewStartTask(repo TaskRepository, runner ScriptRunner) *StartTask {
	return &StartTask{repo: repo, runner: runner}
}

// A setup left half done breaks the task, so it runs to the end even when the page that asked goes away.
func (uc *StartTask) Execute(ctx context.Context, id string) (labs.ScriptResult, error) {
	task, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return labs.ScriptResult{}, err
	}

	if err != nil {
		return labs.ScriptResult{}, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), setupTimeout)
	defer cancel()

	result, err := uc.runner.RunScript(ctx, task.Host, task.Setup)
	switch {
	case errors.Is(err, labs.ErrNotReady):
		return labs.ScriptResult{}, err
	case err != nil:
		return labs.ScriptResult{}, fmt.Errorf("%w: set up %s: %w", labs.ErrUnavailable, id, err)
	}

	return result, nil
}
