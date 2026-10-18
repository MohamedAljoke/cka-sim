package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

const checkTimeout = 2 * time.Minute

// Checked keeps the check's run beside its grade: a check that printed no PASS or FAIL line broke, and its
// output says why.
type Checked struct {
	Grade  tasks.Grade
	Script labs.ScriptResult
}

type CheckTask struct {
	repo   TaskRepository
	runner ScriptRunner
}

func NewCheckTask(repo TaskRepository, runner ScriptRunner) *CheckTask {
	return &CheckTask{repo: repo, runner: runner}
}

func (uc *CheckTask) Execute(ctx context.Context, owner, id string) (Checked, error) {
	task, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return Checked{}, err
	}

	if err != nil {
		return Checked{}, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	result, err := uc.runner.RunScript(ctx, owner, task.Host, task.Check)
	switch {
	case errors.Is(err, labs.ErrNotReady):
		return Checked{}, err
	case err != nil:
		return Checked{}, fmt.Errorf("%w: check %s: %w", labs.ErrUnavailable, id, err)
	}

	return Checked{Grade: tasks.ReadGrade(result.Output), Script: result}, nil
}
