package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/evals"
)

// Lab is the platform as the student's page sees it: it sets a task up, grades it, solves it and tidies it.
type Lab interface {
	Start(ctx context.Context, taskID string) (evals.Script, error)
	Check(ctx context.Context, taskID string) ([]evals.Check, error)
	// Solution is the task's explain.md, the reference the judge compares the tutor with.
	Solution(ctx context.Context, taskID string) (string, error)
	Solve(ctx context.Context, taskID string) (evals.Script, error)
	Tidy(ctx context.Context, taskID string) error
}
