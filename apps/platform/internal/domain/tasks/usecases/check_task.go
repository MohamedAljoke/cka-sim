package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

// CheckTask is a practice check: it grades the task and saves the attempt. An exam grades with GradeTask, since
// its own record already keeps every grade.
type CheckTask struct {
	grade    *GradeTask
	attempts AttemptRepository
	now      func() time.Time
}

func NewCheckTask(grade *GradeTask, attempts AttemptRepository, now func() time.Time) *CheckTask {
	return &CheckTask{grade: grade, attempts: attempts, now: now}
}

// A check that printed no PASS or FAIL line says nothing about the student, so it isn't saved. A save that fails
// doesn't take the grade away from the student.
func (uc *CheckTask) Execute(ctx context.Context, owner, id string) (Checked, error) {
	checked, err := uc.grade.Execute(ctx, owner, id)
	if err != nil || len(checked.Grade.Checks) == 0 {
		return checked, err
	}

	if err := uc.attempts.Append(ctx, owner, tasks.NewAttempt(id, uc.now(), checked.Grade)); err != nil {
		slog.Error("save the attempt", "owner", owner, "task", id, "err", err)
	}

	return checked, nil
}
