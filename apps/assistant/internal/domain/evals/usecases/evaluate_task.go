package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/evals"
)

// EvaluateTask breaks a task, asks the tutor why its check fails, solves it and asks again, then tidies it. The
// grader's failing checks and the task's explain are the ground truth the judge compares the answers with.
type EvaluateTask struct {
	lab     Lab
	tutor   *chatusecases.Reply
	judge   *Judge
	student string
}

func NewEvaluateTask(lab Lab, tutor *chatusecases.Reply, judge *Judge, student string) *EvaluateTask {
	return &EvaluateTask{lab: lab, tutor: tutor, judge: judge, student: student}
}

func (uc *EvaluateTask) Execute(ctx context.Context, taskID string) (result evals.Result) {
	result.TaskID = taskID

	defer func() {
		err := uc.lab.Tidy(context.WithoutCancel(ctx), taskID)
		if err != nil && result.Err == nil {
			result.Err = fmt.Errorf("%w: tidy: %w", evals.ErrPlatform, err)
		}
	}()

	result.Err = uc.run(ctx, taskID, &result)

	return result
}

func (uc *EvaluateTask) run(ctx context.Context, taskID string, result *evals.Result) error {
	failing, err := uc.breakTask(ctx, taskID)
	if err != nil {
		return err
	}

	result.Failing = failing

	solution, err := uc.lab.Solution(ctx, taskID)
	if err != nil {
		return fmt.Errorf("%w: read the solution: %w", evals.ErrPlatform, err)
	}

	if result.Broken, err = uc.ask(ctx, taskID); err != nil {
		return err
	}

	diagnosis, err := uc.judge.Diagnosis(ctx, failing, solution, result.Broken.Text)
	if err != nil {
		return err
	}

	result.Diagnosis = &diagnosis

	if err := uc.solve(ctx, taskID); err != nil {
		return err
	}

	if result.Solved, err = uc.ask(ctx, taskID); err != nil {
		return err
	}

	recheck, err := uc.judge.Recheck(ctx, result.Solved.Text)
	if err != nil {
		return err
	}

	result.Recheck = &recheck

	return nil
}

func (uc *EvaluateTask) breakTask(ctx context.Context, taskID string) ([]evals.Check, error) {
	started, err := uc.lab.Start(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("%w: set up: %w", evals.ErrPlatform, err)
	}

	if !started.Passed {
		return nil, fmt.Errorf("%w: exit code %d: %s", evals.ErrSetupFailed, started.ExitCode, started.Output)
	}

	checks, err := uc.lab.Check(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("%w: check: %w", evals.ErrPlatform, err)
	}

	failing := evals.Failing(checks)
	if len(failing) == 0 {
		return nil, evals.ErrNotBroken
	}

	return failing, nil
}

func (uc *EvaluateTask) solve(ctx context.Context, taskID string) error {
	solved, err := uc.lab.Solve(ctx, taskID)
	if err != nil {
		return fmt.Errorf("%w: solve: %w", evals.ErrPlatform, err)
	}

	if !solved.Passed {
		return fmt.Errorf("%w: exit code %d: %s", evals.ErrNotSolved, solved.ExitCode, solved.Output)
	}

	checks, err := uc.lab.Check(ctx, taskID)
	if err != nil {
		return fmt.Errorf("%w: check: %w", evals.ErrPlatform, err)
	}

	if failing := evals.Failing(checks); len(checks) == 0 || len(failing) > 0 {
		return fmt.Errorf("%w: still failing: %s", evals.ErrNotSolved, descriptions(failing))
	}

	return nil
}

func (uc *EvaluateTask) ask(ctx context.Context, taskID string) (evals.Answer, error) {
	var (
		answer evals.Answer
		text   strings.Builder
	)

	question := []chat.Message{{Role: chat.Student, Text: evals.Question}}

	err := uc.tutor.Execute(ctx, uc.student, taskID, question, func(e chat.Event) error {
		if e.Tool != "" {
			answer.Tools = append(answer.Tools, e.Tool)
		}

		text.WriteString(e.Text)

		return nil
	})
	if err != nil {
		return evals.Answer{}, fmt.Errorf("%w: %w", evals.ErrTutor, err)
	}

	answer.Text = text.String()
	if strings.TrimSpace(answer.Text) == "" {
		return evals.Answer{}, fmt.Errorf("%w: %w", evals.ErrTutor, errors.New("an empty answer"))
	}

	return answer, nil
}

func descriptions(checks []evals.Check) string {
	quoted := make([]string, len(checks))
	for i, c := range checks {
		quoted[i] = fmt.Sprintf("%q", c.Description)
	}

	return strings.Join(quoted, ", ")
}
