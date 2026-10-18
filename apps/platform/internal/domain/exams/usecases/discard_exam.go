package usecases

import (
	"context"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type DiscardExam struct {
	ss *sessions
}

func (uc *DiscardExam) Execute(ctx context.Context, owner string) error {
	s, err := uc.ss.of(owner)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.endIfExpired(ctx)

	switch {
	case s.exam == nil:
		return exams.ErrNoExam
	case !s.exam.Over():
		return exams.ErrRunning
	case s.exam.Scoring():
		return exams.ErrScoring
	}

	if err := s.repo.Clear(s.owner); err != nil {
		return fmt.Errorf("%w: %w", exams.ErrUnavailable, err)
	}

	s.exam = nil

	return nil
}
