package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
)

const (
	maxRounds     = 8
	answerTimeout = 90 * time.Second
)

// Reply answers a student: the model and the tools take turns until the model answers without a tool.
type Reply struct {
	model Model
	tools Toolbox
}

func NewReply(model Model, tools Toolbox) *Reply {
	return &Reply{model: model, tools: tools}
}

func (uc *Reply) Execute(ctx context.Context, student, taskID string, messages []chat.Message,
	emit func(chat.Event) error,
) error {
	if err := chat.Validate(messages); err != nil {
		return err
	}

	if uc.model == nil {
		return chat.ErrNotConfigured
	}

	ctx, cancel := context.WithTimeout(ctx, answerTimeout)
	defer cancel()

	tools, err := uc.tools.Open(ctx, student)
	if err != nil {
		return fmt.Errorf("open the tools: %w", err)
	}
	defer func() { _ = tools.Close() }()

	req := Request{System: system(tools.Instructions(), taskID), Tools: tools.List()}
	for _, m := range messages {
		req.Turns = append(req.Turns, Turn{Role: m.Role, Text: m.Text})
	}

	start := time.Now()
	r := &rounds{emit: emit}

	for round := 1; round <= maxRounds; round++ {
		answer, err := r.generate(ctx, uc.model, req)
		if err != nil {
			return err
		}

		req.Turns = append(req.Turns, answer)
		if len(answer.Calls) == 0 {
			slog.Info("tutor answered", "student", student, "rounds", round, "took", time.Since(start).Round(time.Millisecond))
			return nil
		}

		results, err := r.call(ctx, tools, answer.Calls)
		if err != nil {
			return err
		}

		req.Turns = append(req.Turns, results)
	}

	return chat.ErrTooManySteps
}

func system(instructions, taskID string) string {
	if taskID == "" {
		return instructions
	}

	return instructions + "\n\nThe student's page shows task " + taskID + ". current_task says which task is " +
		"set up on their cluster, which may be another or none. Questions about the page's task, and explain, " +
		"work with its id either way; its check only means something once they've pressed Set up task."
}

// rounds streams the model's text and tool calls as events, and keeps the texts of different rounds apart.
type rounds struct {
	emit   func(chat.Event) error
	wrote  bool
	failed error
}

func (r *rounds) generate(ctx context.Context, model Model, req Request) (Turn, error) {
	first := true
	answer, err := model.Generate(ctx, req, func(text string) {
		if r.failed != nil || text == "" {
			return
		}

		if first && r.wrote {
			text = "\n\n" + text
		}

		first, r.wrote = false, true
		r.failed = r.emit(chat.Event{Text: text})
	})

	switch {
	case r.failed != nil:
		return Turn{}, r.failed
	case err != nil:
		return Turn{}, fmt.Errorf("%w: %w", chat.ErrModelUnavailable, err)
	case len(answer.Calls) == 0 && answer.Text == "":
		return Turn{}, fmt.Errorf("%w: an empty answer", chat.ErrModelUnavailable)
	}

	return answer, nil
}

func (r *rounds) call(ctx context.Context, tools Tools, calls []ToolCall) (Turn, error) {
	results := Turn{Role: chat.Student}

	for _, call := range calls {
		if err := r.emit(chat.Event{Tool: call.Name}); err != nil {
			return Turn{}, err
		}

		output, failed, err := tools.Call(ctx, call.Name, call.Args)
		if err != nil {
			return Turn{}, fmt.Errorf("call %s: %w", call.Name, err)
		}

		results.Results = append(results.Results,
			ToolResult{CallID: call.ID, Name: call.Name, Output: output, Failed: failed})
	}

	return results, nil
}
