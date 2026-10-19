package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

type GetCurrentTask struct {
	platform Platform
}

func NewGetCurrentTask(platform Platform) *GetCurrentTask {
	return &GetCurrentTask{platform: platform}
}

func (uc *GetCurrentTask) Execute(ctx context.Context, student string) (tutor.Task, error) {
	return uc.platform.CurrentTask(ctx, student)
}
