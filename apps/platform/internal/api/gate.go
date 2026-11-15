package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"
)

// gate lets a call through only for a known student who has no exam running.
type gate struct {
	exams examusecases.UseCases
	// token → student name
	students map[string]string
}

func (g gate) outsideAnExam(ctx context.Context, student string) error {
	if !g.knows(student) {
		return refuse(connect.CodeNotFound, errUnknownStudent, nil)
	}

	exam, err := g.exams.Get.Execute(ctx, student)
	switch {
	case errors.Is(err, exams.ErrNoExam):
		return nil
	case errors.Is(err, exams.ErrUnavailable):
		return refuse(connect.CodeUnavailable, exams.ErrUnavailable, err)
	case err != nil:
		return err
	case !exam.Over():
		return refuse(connect.CodeFailedPrecondition, exams.ErrRunning, nil)
	case exam.Scoring():
		return refuse(connect.CodeFailedPrecondition, exams.ErrScoring, nil)
	}

	return nil
}

// With no students set, the only one is the local student.
func (g gate) knows(student string) bool {
	if len(g.students) == 0 {
		return student == localStudent
	}

	for _, name := range g.students {
		if name == student {
			return true
		}
	}

	return false
}
