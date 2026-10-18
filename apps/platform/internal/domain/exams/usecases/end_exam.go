package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type EndExam struct {
	ss *sessions
}

// Checks run in the background; the exam answers at once and is scored once every check is back.
func (uc *EndExam) Execute(ctx context.Context, owner string) (exams.Exam, error) {
	s, err := uc.ss.of(owner)
	if err != nil {
		return exams.Exam{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.endIfExpired(ctx)

	switch {
	case s.exam == nil:
		return exams.Exam{}, exams.ErrNoExam
	case s.exam.Over():
		return exams.Exam{}, exams.ErrEnded
	case s.exam.SettingUp():
		return exams.Exam{}, exams.ErrSettingUp
	}

	ended := s.exam.Clone()
	ended.Ended = s.now()

	if err := s.save(ended); err != nil {
		return exams.Exam{}, err
	}

	s.exam.Ended = ended.Ended
	s.grade(ctx)

	return s.exam.Clone(), nil
}
