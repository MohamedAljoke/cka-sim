package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam"
)

type GetResult struct {
	platform Platform
}

func NewGetResult(platform Platform) *GetResult {
	return &GetResult{platform: platform}
}

func (uc *GetResult) Execute(ctx context.Context, student string) (exam.Result, error) {
	return uc.platform.ExamResult(ctx, student)
}
