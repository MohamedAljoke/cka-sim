package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

type Tasks interface {
	CurrentTask(ctx context.Context, student string) (tutor.Task, error)
}
