package usecases

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

func TestACheckWithNoTaskGradesTheCurrentOne(t *testing.T) {
	p := &fakePlatform{current: tutor.Task{ID: "wl-scale"}, grade: tutor.Grade{Earned: 1, Total: 2}}

	grade, err := NewCheckTask(p).Execute(context.Background(), "local", "")

	if err != nil || p.checked != "wl-scale" || grade.TaskID != "wl-scale" || grade.Earned != 1 {
		t.Errorf("got %+v, %v; checked %q", grade, err, p.checked)
	}
}

func TestACheckOfANamedTaskGradesThatOne(t *testing.T) {
	p := &fakePlatform{current: tutor.Task{ID: "wl-scale"}}

	grade, err := NewCheckTask(p).Execute(context.Background(), "local", "ar-aggregate-view")

	if err != nil || p.checked != "ar-aggregate-view" || grade.TaskID != "ar-aggregate-view" {
		t.Errorf("got %+v, %v; checked %q", grade, err, p.checked)
	}
}

func TestACheckWithNoTaskAndNoneSetUpIsRefused(t *testing.T) {
	p := &fakePlatform{currentErr: fmt.Errorf("%w: no_current_task", tutor.ErrRefused)}

	_, err := NewCheckTask(p).Execute(context.Background(), "local", "")

	if !errors.Is(err, tutor.ErrRefused) || p.checked != "" {
		t.Errorf("got %v; checked %q", err, p.checked)
	}
}

func TestTheSolutionWithNoTaskIsTheCurrentOnes(t *testing.T) {
	p := &fakePlatform{current: tutor.Task{ID: "wl-scale"}, solutions: map[string]string{"wl-scale": "Run kubectl scale."}}

	solution, err := NewGetSolution(p).Execute(context.Background(), "local", "")

	if err != nil || solution != "Run kubectl scale." {
		t.Errorf("got %q, %v", solution, err)
	}
}

type fakePlatform struct {
	current    tutor.Task
	currentErr error
	grade      tutor.Grade
	checked    string
	solutions  map[string]string
}

func (p *fakePlatform) CurrentTask(context.Context, string) (tutor.Task, error) {
	return p.current, p.currentErr
}

func (p *fakePlatform) Check(_ context.Context, _, taskID string) (tutor.Grade, error) {
	p.checked = taskID
	return p.grade, nil
}

func (p *fakePlatform) Inspect(context.Context, string, []string) (tutor.Inspection, error) {
	return tutor.Inspection{}, nil
}

func (p *fakePlatform) Solution(_ context.Context, _, taskID string) (string, error) {
	return p.solutions[taskID], nil
}
