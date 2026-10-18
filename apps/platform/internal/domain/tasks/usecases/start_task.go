package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type StartTask struct {
	repo  TaskRepository
	heal  *HealCluster
	setUp *SetUpTask
}

func NewStartTask(repo TaskRepository, heal *HealCluster, setUp *SetUpTask) *StartTask {
	return &StartTask{repo: repo, heal: heal, setUp: setUp}
}

// Execute heals what earlier tasks broke before setting this one up, so a task left half solved can't break it. A
// reset that fails answers in place of the setup.
func (uc *StartTask) Execute(ctx context.Context, owner, id string) (labs.ScriptResult, error) {
	_, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return labs.ScriptResult{}, err
	}

	if err != nil {
		return labs.ScriptResult{}, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	healed, err := uc.heal.Execute(ctx, owner)
	if err != nil || !healed.Passed() {
		return healed, err
	}

	return uc.setUp.Execute(ctx, owner, id)
}
