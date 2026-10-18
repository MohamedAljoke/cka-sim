package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type SelftestTask struct {
	repo    TaskRepository
	runner  ScriptRunner
	cluster *cluster
	start   *StartTask
	check   *CheckTask
}

func NewSelftestTask(repo TaskRepository, runner ScriptRunner, cluster *cluster, start *StartTask,
	check *CheckTask,
) *SelftestTask {
	return &SelftestTask{repo: repo, runner: runner, cluster: cluster, start: start, check: check}
}

// Execute proves a task is fair on the running lab: after its setup nothing earns points, and after its
// solution everything does.
func (uc *SelftestTask) Execute(ctx context.Context, id string) error {
	task, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return err
	}

	if err != nil {
		return fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	started, err := uc.start.Execute(ctx, id)
	if err != nil {
		return err
	}

	if !started.Passed() {
		return fmt.Errorf("set up %s: exit code %d:\n%s", id, started.ExitCode, started.Output)
	}

	before, err := uc.check.Execute(ctx, id)
	if err != nil {
		return err
	}

	switch {
	case before.Grade.Total == 0:
		return fmt.Errorf("check %s prints no checks:\n%s", id, before.Script.Output)
	case before.Grade.Earned != 0:
		return fmt.Errorf("%s earns %d/%d before the solution: %s", id, before.Grade.Earned, before.Grade.Total,
			checks(before.Grade, true))
	}

	if err := uc.solve(ctx, task); err != nil {
		return err
	}

	after, err := uc.check.Execute(ctx, id)
	if err != nil {
		return err
	}

	if after.Grade.Total == 0 || after.Grade.Earned != after.Grade.Total {
		return fmt.Errorf("%s earns %d/%d after the solution: %s", id, after.Grade.Earned, after.Grade.Total,
			checks(after.Grade, false))
	}

	return nil
}

func (uc *SelftestTask) solve(ctx context.Context, task tasks.Task) error {
	ctx, cancel := context.WithTimeout(ctx, setupTimeout)
	defer cancel()

	uc.cluster.scripts.RLock()
	defer uc.cluster.scripts.RUnlock()

	result, err := uc.runner.RunScript(ctx, task.Host, task.Solve)
	switch {
	case errors.Is(err, labs.ErrNotReady):
		return err
	case err != nil:
		return fmt.Errorf("%w: solve %s: %w", labs.ErrUnavailable, task.ID, err)
	case !result.Passed():
		return fmt.Errorf("solve %s: exit code %d:\n%s", task.ID, result.ExitCode, result.Output)
	}

	return nil
}

func checks(grade tasks.Grade, passed bool) string {
	var names []string

	for _, c := range grade.Checks {
		if c.Passed == passed {
			names = append(names, fmt.Sprintf("%q", c.Description))
		}
	}

	return strings.Join(names, ", ")
}
