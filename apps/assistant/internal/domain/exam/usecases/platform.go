package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam"
)

type Platform interface {
	StartExam(ctx context.Context, student string, count, minutes int, selection exam.Selection) (exam.Exam, error)
	ExamResult(ctx context.Context, student string) (exam.Result, error)
}
