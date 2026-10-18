package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	taskusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

type BeginExam struct {
	ss   *sessions
	list *taskusecases.ListTasks
	lab  *labusecases.GetLab
}

// Setups run in the background; the exam answers at once and its tasks report as each setup lands.
func (uc *BeginExam) Execute(ctx context.Context, owner string, count, minutes int) (exams.Exam, error) {
	if err := exams.Validate(count, minutes); err != nil {
		return exams.Exam{}, err
	}

	lab, err := uc.lab.Execute(ctx, owner)
	if err != nil {
		return exams.Exam{}, err
	}

	if lab.State != labs.Ready {
		return exams.Exam{}, labs.ErrNotReady
	}

	all, err := uc.list.Execute(ctx, tasks.Filter{})
	if err != nil {
		return exams.Exam{}, err
	}

	s, err := uc.ss.of(owner)
	if err != nil {
		return exams.Exam{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.endIfExpired(ctx)

	switch {
	case s.exam == nil:
	case !s.exam.Over():
		return exams.Exam{}, exams.ErrRunning
	case s.exam.Scoring():
		return exams.Exam{}, exams.ErrScoring
	}

	drawn := tasks.Draw(all, count, s.rng)
	if len(drawn) == 0 {
		return exams.Exam{}, exams.ErrNoTasks
	}

	exam := exams.New(drawn, minutes)
	if err := s.save(exam); err != nil {
		return exams.Exam{}, err
	}

	s.exam = &exam
	s.setUp(ctx)

	return exam.Clone(), nil
}
