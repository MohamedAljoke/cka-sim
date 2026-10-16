package usecases

import (
	"context"
	"slices"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type fakeRepository struct {
	tasks []tasks.Task
	err   error
}

func (r fakeRepository) List(context.Context) ([]tasks.Task, error) { return r.tasks, r.err }

func (r fakeRepository) Get(_ context.Context, id string) (tasks.Task, error) {
	if r.err != nil {
		return tasks.Task{}, r.err
	}
	i := slices.IndexFunc(r.tasks, func(t tasks.Task) bool { return t.ID == id })
	if i < 0 {
		return tasks.Task{}, tasks.ErrNotFound
	}
	return r.tasks[i], nil
}
