package usecases

import (
	"context"
	"errors"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

// Reply answers a student. Until the AI is plugged in it's a stand-in that only reads the current task, so the
// whole path from the page to the platform's tools can be tried without a model.
type Reply struct {
	tasks Tasks
}

func NewReply(tasks Tasks) *Reply {
	return &Reply{tasks: tasks}
}

func (uc *Reply) Execute(ctx context.Context, student string, messages []chat.Message,
	emit func(chat.Event) error,
) error {
	if err := chat.Validate(messages); err != nil {
		return err
	}

	if err := emit(chat.Event{Tool: "current_task"}); err != nil {
		return err
	}

	where := "You have no task set up on your cluster yet."
	task, err := uc.tasks.CurrentTask(ctx, student)

	switch {
	case err == nil:
		where = "You're working on “" + task.Title + "”."
	case !errors.Is(err, tutor.ErrRefused):
		return err
	}

	for _, piece := range []string{
		"I'm a stand-in until the AI is plugged in. ",
		where,
		" You said: “" + messages[len(messages)-1].Text + "”",
	} {
		if err := emit(chat.Event{Text: piece}); err != nil {
			return err
		}
	}

	return nil
}
