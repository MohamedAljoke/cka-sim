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
	s    *session
	list *taskusecases.ListTasks
	lab  *labusecases.GetLab
}

// Setups run in the background; the exam answers at once and its tasks report as each setup lands.
func (uc *BeginExam) Execute(ctx context.Context, count, minutes int) (exams.Exam, error) {
	if err := exams.Validate(count, minutes); err != nil {
		return exams.Exam{}, err
	}

	lab, err := uc.lab.Execute(ctx)
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

	uc.s.mu.Lock()
	defer uc.s.mu.Unlock()

	uc.s.endIfExpired(ctx)

	switch {
	case uc.s.exam == nil:
	case !uc.s.exam.Over():
		return exams.Exam{}, exams.ErrRunning
	case uc.s.exam.Scoring():
		return exams.Exam{}, exams.ErrScoring
	}

	drawn := tasks.Draw(all, count, uc.s.rng)
	if len(drawn) == 0 {
		return exams.Exam{}, exams.ErrNoTasks
	}

	exam := exams.New(drawn, minutes)
	if err := uc.s.save(exam); err != nil {
		return exams.Exam{}, err
	}

	uc.s.exam = &exam
	uc.s.setUp(ctx)

	return exam.Clone(), nil
}
