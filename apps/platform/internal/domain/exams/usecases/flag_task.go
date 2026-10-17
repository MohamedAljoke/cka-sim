package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type FlagTask struct {
	s *session
}

func (uc *FlagTask) Execute(ctx context.Context, id string, flagged bool) error {
	uc.s.mu.Lock()
	defer uc.s.mu.Unlock()

	uc.s.endIfExpired(ctx)

	switch {
	case uc.s.exam == nil:
		return exams.ErrNoExam
	case uc.s.exam.Over():
		return exams.ErrEnded
	}

	i := uc.s.exam.Index(id)
	if i < 0 {
		return exams.ErrNotInExam
	}

	changed := uc.s.exam.Clone()
	changed.Tasks[i].Flagged = flagged

	if err := uc.s.save(changed); err != nil {
		return err
	}

	uc.s.exam.Tasks[i].Flagged = flagged

	return nil
}
