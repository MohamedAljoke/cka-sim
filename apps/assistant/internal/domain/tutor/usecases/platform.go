package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

// Platform is the read-only view of a student's task and cluster that cka-sim gives the assistant.
type Platform interface {
	CurrentTask(ctx context.Context, student string) (tutor.Task, error)
	Check(ctx context.Context, student, taskID string) (tutor.Grade, error)
	Inspect(ctx context.Context, student string, args []string) (tutor.Inspection, error)
	Solution(ctx context.Context, student, taskID string) (string, error)
}
