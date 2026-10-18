package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type ResumeExam struct {
	ss *sessions
}

// Execute picks up the setups or checks that a stopped server left unfinished, for every owner with an exam.
func (uc *ResumeExam) Execute(ctx context.Context) error {
	owners, err := uc.ss.repo.Owners()
	if err != nil {
		return fmt.Errorf("%w: list the saved exams: %w", exams.ErrUnavailable, err)
	}

	var errs []error

	for _, owner := range owners {
		s, err := uc.ss.of(owner)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		s.resume(ctx)
	}

	return errors.Join(errs...)
}

func (s *session) resume(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch {
	case s.exam == nil:
	case s.exam.Scoring():
		s.grade(ctx)
	case s.exam.SettingUp():
		s.setUp(ctx)
	default:
		s.endIfExpired(ctx)
	}
}
