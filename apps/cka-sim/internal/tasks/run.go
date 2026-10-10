package tasks

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/grader"
)

type Runner interface {
	Run(ctx context.Context, host, taskID string, script []byte) (string, error)
}

const scriptTimeout = 3 * time.Minute

func Start(ctx context.Context, fsys fs.FS, r Runner, t Task) error {
	if _, err := run(ctx, fsys, r, t, "setup.sh"); err != nil {
		return fmt.Errorf("set up %s: %w", t.ID, err)
	}
	return nil
}

func Check(ctx context.Context, fsys fs.FS, r Runner, t Task) (grader.Result, error) {
	out, err := run(ctx, fsys, r, t, "check.sh")
	if err != nil {
		return grader.Result{}, fmt.Errorf("check %s: %w", t.ID, err)
	}
	return grader.Parse(out), nil
}

func Solve(ctx context.Context, fsys fs.FS, r Runner, t Task) error {
	if _, err := run(ctx, fsys, r, t, "solution.sh"); err != nil {
		return fmt.Errorf("solve %s: %w", t.ID, err)
	}
	return nil
}

// Heal runs every reset.sh among all, one at a time, so a task left unsolved can't break the
// next one. A failed reset doesn't stop the others.
func Heal(ctx context.Context, fsys fs.FS, r Runner, all []Task) error {
	var errs []error
	for _, t := range all {
		if !t.Reset {
			continue
		}
		if _, err := run(ctx, fsys, r, t, ResetScript); err != nil {
			errs = append(errs, fmt.Errorf("reset %s: %w", t.ID, err))
		}
	}
	return errors.Join(errs...)
}

// Selftest proves a task is fair: setup leaves nothing to earn, and the solution earns everything.
func Selftest(ctx context.Context, fsys fs.FS, r Runner, t Task) error {
	if err := Heal(ctx, fsys, r, []Task{t}); err != nil {
		return err
	}
	if err := Start(ctx, fsys, r, t); err != nil {
		return err
	}
	before, err := Check(ctx, fsys, r, t)
	if err != nil {
		return err
	}
	if before.Total == 0 {
		return fmt.Errorf("%s: check.sh prints no checks", t.ID)
	}
	if before.Earned != 0 {
		return fmt.Errorf("%s: earns %d/%d before the solution: %s", t.ID, before.Earned, before.Total, describe(before, true))
	}
	if err := Solve(ctx, fsys, r, t); err != nil {
		return err
	}
	after, err := Check(ctx, fsys, r, t)
	if err != nil {
		return err
	}
	if after.Earned != after.Total {
		return fmt.Errorf("%s: earns %d/%d after the solution: %s", t.ID, after.Earned, after.Total, describe(after, false))
	}
	return nil
}

func run(ctx context.Context, fsys fs.FS, r Runner, t Task, script string) (string, error) {
	body, err := fs.ReadFile(fsys, path.Join(t.Dir, script))
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, scriptTimeout)
	defer cancel()
	return r.Run(ctx, t.Host, t.ID, body)
}

func describe(r grader.Result, passed bool) string {
	var names []string
	for _, c := range r.Checks {
		if c.Passed == passed {
			names = append(names, fmt.Sprintf("%q", c.Description))
		}
	}
	return strings.Join(names, ", ")
}
