package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type GetExam struct {
	ss *sessions
}

// Reading an exam past its deadline ends it, so the server decides when time is up, not the browser.
func (uc *GetExam) Execute(ctx context.Context, owner string) (exams.Exam, error) {
	s, err := uc.ss.of(owner)
	if err != nil {
		return exams.Exam{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.endIfExpired(ctx)

	if s.exam == nil {
		return exams.Exam{}, exams.ErrNoExam
	}

	return s.exam.Clone(), nil
}
