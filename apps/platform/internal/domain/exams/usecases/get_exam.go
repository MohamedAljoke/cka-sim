package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type GetExam struct {
	s *session
}

// Reading an exam past its deadline ends it, so the server decides when time is up, not the browser.
func (uc *GetExam) Execute(ctx context.Context) (exams.Exam, error) {
	uc.s.mu.Lock()
	defer uc.s.mu.Unlock()

	uc.s.endIfExpired(ctx)

	if uc.s.exam == nil {
		return exams.Exam{}, exams.ErrNoExam
	}

	return uc.s.exam.Clone(), nil
}
