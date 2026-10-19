package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

type CheckTask struct {
	platform Platform
}

func NewCheckTask(platform Platform) *CheckTask {
	return &CheckTask{platform: platform}
}

// Execute grades taskID, or the task the student is on when taskID is empty.
func (uc *CheckTask) Execute(ctx context.Context, student, taskID string) (tutor.Grade, error) {
	taskID, err := orCurrent(ctx, uc.platform, student, taskID)
	if err != nil {
		return tutor.Grade{}, err
	}

	grade, err := uc.platform.Check(ctx, student, taskID)
	grade.TaskID = taskID

	return grade, err
}

func orCurrent(ctx context.Context, platform Platform, student, taskID string) (string, error) {
	if taskID != "" {
		return taskID, nil
	}

	task, err := platform.CurrentTask(ctx, student)

	return task.ID, err
}
