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
	row = Row{Task: a.Task.ID, Attempt: a.Name, Marks: unscored, Review: Review{Tools: []string{}}}
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
	return row, nil
}

func (h Harness) review(ctx context.Context) (Review, error) {
	if err := h.Platform.Connect(ctx); err != nil {
		return Review{Tools: []string{}}, err
	}
	script, err := reviewScript()
	if err != nil {
		return Review{Tools: []string{}}, err
	}
	r, err := h.ask(ctx, script)
	if errors.Is(err, ErrNoMCP) {
		r, err = h.ask(ctx, script)
	}
	return r, err
}

func (h Harness) ask(ctx context.Context, script string) (Review, error) {
	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()
	out, runErr := h.Lab.Run(ctx, labagent.BaseHost, "eval", []byte(script))
	r, err := Parse(out)
	if err != nil && runErr != nil && !errors.Is(err, ErrNotLoggedIn) && !errors.Is(err, ErrNoMCP) {
		return r, runErr
	}
	return r, err
}
