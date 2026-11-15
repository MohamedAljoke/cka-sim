package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

type GetExamResult struct {
	ss *sessions
}

// Only a graded exam has a result, so a review can never see a task while the exam can still be solved.
func (uc *GetExamResult) Execute(ctx context.Context, owner string) (exams.Exam, error) {
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
	case !s.exam.Over():
		return exams.Exam{}, exams.ErrRunning
	case s.exam.Scoring():
		return exams.Exam{}, exams.ErrScoring
	}

	return s.exam.Clone(), nil
}
