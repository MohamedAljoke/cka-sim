package eval

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"

	"sigs.k8s.io/yaml"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

const (
	Solved    = "solved"
	Untouched = "untouched"
	setFile   = "evals.yaml"
)

type Outcome string

const (
	EarnsAll  Outcome = "pass"
	EarnsSome Outcome = "partial"
	EarnsNone Outcome = "fail"
)

type Attempt struct {
	Task     tasks.Task `json:"-"`
	Name     string     `json:"name"`
	Script   string     `json:"script,omitempty"`
	Expect   Outcome    `json:"expect,omitempty"`
	Mentions []string   `json:"mentions,omitempty"`
}

func Load(fsys fs.FS, all []tasks.Task) ([]Attempt, error) {
	var set []Attempt
	for _, t := range all {
		file := path.Join(t.Dir, setFile)
		data, err := fs.ReadFile(fsys, file)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var attempts []Attempt
		if err := yaml.UnmarshalStrict(data, &attempts); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		var names []string
		for _, a := range attempts {
			if err := a.validate(names); err != nil {
				return nil, fmt.Errorf("%s: %w", file, err)
			}
			names = append(names, a.Name)
			a.Task = t
			set = append(set, a)
		}
	}
	return set, nil
}

func Only(set []Attempt, task string) []Attempt {
	return slices.DeleteFunc(slices.Clone(set), func(a Attempt) bool { return a.Task.ID != task })
}

func (a Attempt) want() Outcome {
	switch a.Name {
	case Solved:
		return EarnsAll
	case Untouched:
		return EarnsNone
	}
	if a.Expect == "" {
		return EarnsSome
	}
	return a.Expect
}

func (a Attempt) validate(seen []string) error {
	builtIn := a.Name == Solved || a.Name == Untouched
	switch {
	case a.Name == "":
		return errors.New("an attempt has no name")
	case slices.Contains(seen, a.Name):
		return fmt.Errorf("attempt %q is listed twice", a.Name)
	case builtIn && a.Script != "":
		return fmt.Errorf("attempt %q is built in and takes no script", a.Name)
	case !builtIn && a.Script == "":
		return fmt.Errorf("attempt %q needs a script", a.Name)
	case builtIn && a.Expect != "":
		return fmt.Errorf("attempt %q is built in, so its outcome is known", a.Name)
	case !slices.Contains([]Outcome{"", EarnsAll, EarnsSome, EarnsNone}, a.Expect):
		return fmt.Errorf("attempt %q expects %q, want pass, partial or fail", a.Name, a.Expect)
	}
	return nil
}

func (a Attempt) play(ctx context.Context, fsys fs.FS, lab tasks.Runner) error {
	switch a.Name {
	case Untouched:
		return nil
	case Solved:
		return tasks.Solve(ctx, fsys, lab, a.Task)
	}
	if _, err := tasks.Attempt(ctx, fsys, lab, a.Task, []byte(a.Script)); err != nil {
		return fmt.Errorf("play %s/%s: %w", a.Task.ID, a.Name, err)
	}
	return nil
}
