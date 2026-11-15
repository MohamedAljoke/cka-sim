package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam"
)

type StartExam struct {
	platform Platform
}

func NewStartExam(platform Platform) *StartExam {
	return &StartExam{platform: platform}
}

func (uc *StartExam) Execute(ctx context.Context, student string, count, minutes int, selection exam.Selection,
) (exam.Exam, error) {
	return uc.platform.StartExam(ctx, student, count, minutes, selection)
}
