package usecases

import (
	"context"
	"fmt"
	"slices"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

const maxHistory = 200

type GetHistory struct {
	attempts AttemptRepository
}

func NewGetHistory(attempts AttemptRepository) *GetHistory {
	return &GetHistory{attempts: attempts}
}

// Execute answers the owner's attempts and notes newest first, only those of taskIDs unless it's empty.
func (uc *GetHistory) Execute(ctx context.Context, owner string, taskIDs []string) ([]tasks.Entry, error) {
	all, err := uc.attempts.List(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", tasks.ErrNoHistory, err)
	}

	kept := []tasks.Entry{}

	for _, entry := range slices.Backward(all) {
		if len(kept) == maxHistory {
			break
		}

		if len(taskIDs) == 0 || slices.Contains(taskIDs, entry.TaskID) {
			kept = append(kept, entry)
		}
	}

	return kept, nil
}
