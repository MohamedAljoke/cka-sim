package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type EndExam struct {
	s *session
}

// Checks run in the background; the exam answers at once and is scored once every check is back.
func (uc *EndExam) Execute(ctx context.Context) (exams.Exam, error) {
	uc.s.mu.Lock()
	defer uc.s.mu.Unlock()

	uc.s.endIfExpired(ctx)

	switch {
	case uc.s.exam == nil:
		return exams.Exam{}, exams.ErrNoExam
	case uc.s.exam.Over():
		return exams.Exam{}, exams.ErrEnded
	case uc.s.exam.SettingUp():
		return exams.Exam{}, exams.ErrSettingUp
	}

	ended := uc.s.exam.Clone()
	ended.Ended = uc.s.now()

	if err := uc.s.save(ended); err != nil {
		return exams.Exam{}, err
	}

	uc.s.exam.Ended = ended.Ended
	uc.s.grade(ctx)

	return uc.s.exam.Clone(), nil
}
