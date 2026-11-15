package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

// MoveToTask starts a task for a caller that, unlike the page, doesn't tidy the task it leaves: the task the
// owner is on is tidied first when it's another one.
type MoveToTask struct {
	repo     TaskRepository
	tidy     *TidyTask
	start    *StartTask
	clusters *clusters
}

func NewMoveToTask(repo TaskRepository, tidy *TidyTask, start *StartTask, clusters *clusters) *MoveToTask {
	return &MoveToTask{repo: repo, tidy: tidy, start: start, clusters: clusters}
}

// A tidy whose reset or namespace delete fails doesn't stop the move: the task stays touched, so the start's heal
// runs its reset again and answers in place of the setup if it still fails.
func (uc *MoveToTask) Execute(ctx context.Context, owner, id string) (tasks.Task, labs.ScriptResult, error) {
	task, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return tasks.Task{}, labs.ScriptResult{}, err
	}

	if err != nil {
		return tasks.Task{}, labs.ScriptResult{}, fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	if current := uc.clusters.of(owner).currentTask(); current != "" && current != id {
		err := uc.tidy.Execute(ctx, owner, current)
		switch {
		case errors.Is(err, labs.ErrNotReady), errors.Is(err, labs.ErrUnavailable),
			errors.Is(err, tasks.ErrNotFound), errors.Is(err, tasks.ErrUnavailable):
			return tasks.Task{}, labs.ScriptResult{}, err
		case err != nil:
			slog.Warn("tidy the task left", "owner", owner, "task", current, "err", err)
		}
	}

	started, err := uc.start.Execute(ctx, owner, id)
	if err != nil {
		return tasks.Task{}, labs.ScriptResult{}, err
	}

	return task, started, nil
}
