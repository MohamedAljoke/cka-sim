package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type SaveNote struct {
	repo     TaskRepository
	attempts AttemptRepository
	now      func() time.Time
}

func NewSaveNote(repo TaskRepository, attempts AttemptRepository, now func() time.Time) *SaveNote {
	return &SaveNote{repo: repo, attempts: attempts, now: now}
}

func (uc *SaveNote) Execute(ctx context.Context, owner, id, text string) error {
	_, err := uc.repo.Get(ctx, id)
	if errors.Is(err, tasks.ErrNotFound) {
		return err
	}

	if err != nil {
		return fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)
	}

	note, err := tasks.NewNote(id, uc.now(), text)
	if err != nil {
		return err
	}

	if err := uc.attempts.Append(ctx, owner, note); err != nil {
		return fmt.Errorf("%w: %w", tasks.ErrNoHistory, err)
	}

	return nil
}
