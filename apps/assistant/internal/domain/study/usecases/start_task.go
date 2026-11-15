package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
)

type StartTask struct {
	platform Platform
}

func NewStartTask(platform Platform) *StartTask {
	return &StartTask{platform: platform}
}

func (uc *StartTask) Execute(ctx context.Context, student, taskID string) (study.Started, error) {
	return uc.platform.StartTask(ctx, student, taskID)
}
