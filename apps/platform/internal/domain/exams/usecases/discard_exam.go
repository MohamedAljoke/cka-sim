package usecases

import (
	"context"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type DiscardExam struct {
	s *session
}

func (uc *DiscardExam) Execute(ctx context.Context) error {
	uc.s.mu.Lock()
	defer uc.s.mu.Unlock()

	uc.s.endIfExpired(ctx)

	switch {
	case uc.s.exam == nil:
		return exams.ErrNoExam
	case !uc.s.exam.Over():
		return exams.ErrRunning
	case uc.s.exam.Scoring():
		return exams.ErrScoring
	}

	if err := uc.s.repo.Clear(); err != nil {
		return fmt.Errorf("%w: %w", exams.ErrUnavailable, err)
	}

	uc.s.exam = nil

	return nil
}
