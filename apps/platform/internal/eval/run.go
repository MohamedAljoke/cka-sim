package eval

import (
	"context"
	"errors"
	"io/fs"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/grader"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

type Platform interface {
	Start(ctx context.Context, task string) error
	Check(ctx context.Context, task string) (grader.Result, error)
	Connect(ctx context.Context) error
	Tidy(ctx context.Context, task string) error
}

type Harness struct {
	Files    fs.FS
	Platform Platform
	Lab      tasks.Runner
	Timeout  time.Duration
}

func (h Harness) Run(ctx context.Context, set []Attempt, done func(Row)) ([]Row, error) {
	var rows []Row
	for _, a := range set {
		row, err := h.one(ctx, a)
		rows = append(rows, row)
		done(row)
		if errors.Is(err, ErrNotLoggedIn) || ctx.Err() != nil {
			return rows, errors.Join(err, ctx.Err())
		}
	}
	return rows, nil
}

func (h Harness) one(ctx context.Context, a Attempt) (row Row, err error) {
	row = Row{Task: a.Task.ID, Attempt: a.Name, Marks: unscored, Review: Review{Session: Session{Tools: []string{}}}}
	defer func() {
		if err != nil {
			row.Error = err.Error()
		}
	}()
	if err := h.Platform.Start(ctx, a.Task.ID); err != nil {
		return row, err
	}
	defer func() {
		if tidyErr := h.Platform.Tidy(context.WithoutCancel(ctx), a.Task.ID); tidyErr != nil && err == nil {
			err = tidyErr
		}
	}()
	if err := a.play(ctx, h.Files, h.Lab); err != nil {
		return row, err
	}
	if row.Truth, err = h.Platform.Check(ctx, a.Task.ID); err != nil {
		return row, err
	}
	if row.Invalid = invalid(a, row.Truth); row.Invalid != "" {
		return row, nil
	}
	if row.Review, err = h.review(ctx); err != nil {
		return row, err
	}
	row.Marks = Score(a, row.Truth, row.Review)
	if row.Review.Verdict == nil {
		return row, nil
	}
	j, err := h.judge(ctx, a, row.Truth, *row.Review.Verdict)
	if errors.Is(err, ErrNotLoggedIn) {
		return row, err
	}
	if err != nil {
		row.JudgeNote = err.Error()
		return row, nil
	}
	row.Judgment, row.Marks = &j, j.marks(a, row.Marks)
	return row, nil
}

func (h Harness) review(ctx context.Context) (Review, error) {
	empty := Review{Session: Session{Tools: []string{}}}
	if err := h.Platform.Connect(ctx); err != nil {
		return empty, err
	}
	script, err := claudeScript(prompt, verdictSchema)
	if err != nil {
		return empty, err
	}
	r, err := h.askReview(ctx, script)
	if errors.Is(err, ErrNoMCP) {
		r, err = h.askReview(ctx, script)
	}
	return r, err
}

func (h Harness) askReview(ctx context.Context, script string) (Review, error) {
	out, runErr := h.claude(ctx, script)
	r, err := ParseReview(out)
	return r, cause(err, runErr)
}

func (h Harness) judge(ctx context.Context, a Attempt, truth grader.Result, v Verdict) (Judgment, error) {
	var j Judgment
	question, err := judgePrompt(a, truth, v)
	if err != nil {
		return j, err
	}
	script, err := claudeScript(question, judgeSchema)
	if err != nil {
		return j, err
	}
	out, runErr := h.claude(ctx, script)
	s, err := Parse(out, &j)
	if err := cause(err, runErr); err != nil {
		return j, err
	}
	if s.Failure != "" {
		return j, errors.New(s.Failure)
	}
	return j, nil
}

func (h Harness) claude(ctx context.Context, script string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()
	return h.Lab.Run(ctx, labagent.BaseHost, "eval", []byte(script))
}

func cause(parseErr, runErr error) error {
	if parseErr != nil && runErr != nil && !errors.Is(parseErr, ErrNotLoggedIn) && !errors.Is(parseErr, ErrNoMCP) {
		return runErr
	}
	return parseErr
}
