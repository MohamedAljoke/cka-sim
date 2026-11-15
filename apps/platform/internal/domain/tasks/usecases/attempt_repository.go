package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

// AttemptRepository keeps each owner's attempts and notes in the order they were added.
type AttemptRepository interface {
	Append(ctx context.Context, owner string, entry tasks.Entry) error
	List(ctx context.Context, owner string) ([]tasks.Entry, error)
}
