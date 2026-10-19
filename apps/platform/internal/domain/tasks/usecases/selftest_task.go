package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type SelftestTask struct {
	start *StartTask
	check *CheckTask
	solve *SolveTask
}

func NewSelftestTask(start *StartTask, check *CheckTask, solve *SolveTask) *SelftestTask {
	return &SelftestTask{start: start, check: check, solve: solve}
}

// Execute proves a task is fair on the running lab: after its setup nothing earns points, and after its
// solution everything does.
func (uc *SelftestTask) Execute(ctx context.Context, owner, id string) error {
	started, err := uc.start.Execute(ctx, owner, id)
	if err != nil {
		return err
	}

	if !started.Passed() {
		return fmt.Errorf("set up %s: exit code %d:\n%s", id, started.ExitCode, started.Output)
	}

	before, err := uc.check.Execute(ctx, owner, id)
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

	solved, err := uc.solve.Execute(ctx, owner, id)
	if err != nil {
		return err
	}

	if !solved.Passed() {
		return fmt.Errorf("solve %s: exit code %d:\n%s", id, solved.ExitCode, solved.Output)
	}

	after, err := uc.check.Execute(ctx, owner, id)
	if err != nil {
		return err
	}

	if after.Grade.Total == 0 || after.Grade.Earned != after.Grade.Total {
		return fmt.Errorf("%s earns %d/%d after the solution: %s", id, after.Grade.Earned, after.Grade.Total,
			checks(after.Grade, false))
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
