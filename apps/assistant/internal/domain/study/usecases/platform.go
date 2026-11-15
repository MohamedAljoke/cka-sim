package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
)

// Platform is what cka-sim gives a study session: the catalog, setting a task up, and the student's record.
type Platform interface {
	ListTasks(ctx context.Context, student string, filter study.Filter) ([]study.TaskSummary, error)
	StartTask(ctx context.Context, student, taskID string) (study.Started, error)
	History(ctx context.Context, student string, taskIDs []string) ([]study.Entry, error)
	SaveNote(ctx context.Context, student, taskID, text string) error
}
