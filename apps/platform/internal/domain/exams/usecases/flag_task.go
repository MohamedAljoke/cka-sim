package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type FlagTask struct {
	ss *sessions
}

func (uc *FlagTask) Execute(ctx context.Context, owner, id string, flagged bool) error {
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
	case s.exam.Over():
		return exams.ErrEnded
	}

	i := s.exam.Index(id)
	if i < 0 {
		return exams.ErrNotInExam
	}

	changed := s.exam.Clone()
	changed.Tasks[i].Flagged = flagged

	if err := s.save(changed); err != nil {
		return err
	}

	s.exam.Tasks[i].Flagged = flagged

	return nil
}
