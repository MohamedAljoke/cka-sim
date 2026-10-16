package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type TaskRepository interface {
	List(ctx context.Context) ([]tasks.Task, error)
	Get(ctx context.Context, id string) (tasks.Task, error)
}
