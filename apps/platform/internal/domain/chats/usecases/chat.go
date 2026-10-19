package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"
)

type Chat struct {
	tutor Tutor
	exam  *examusecases.GetExam
}

func NewChat(tutor Tutor, exam *examusecases.GetExam) *Chat {
	return &Chat{tutor: tutor, exam: exam}
}

// Execute passes a student's conversation to the tutor, except during their exam, where there is no help.
func (uc *Chat) Execute(ctx context.Context, student, taskID string, messages []chats.Message,
	emit func(chats.Event) error,
) error {
	if err := chats.Validate(messages); err != nil {
		return err
	}

	if uc.tutor == nil {
		return chats.ErrOff
	}

	if err := uc.outsideAnExam(ctx, student); err != nil {
		return err
	}

	return uc.tutor.Chat(ctx, student, taskID, messages, emit)
}

func (uc *Chat) outsideAnExam(ctx context.Context, student string) error {
	exam, err := uc.exam.Execute(ctx, student)

	switch {
	case errors.Is(err, exams.ErrNoExam):
		return nil
	case err != nil:
		return fmt.Errorf("%w: %w", exams.ErrUnavailable, err)
	case !exam.Over():
		return exams.ErrRunning
	case exam.Scoring():
		return exams.ErrScoring
	}

	return nil
}
