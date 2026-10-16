package usecases

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestListTasksSortsByID(t *testing.T) {
	uc := NewListTasks(fakeRepository{tasks: []tasks.Task{{ID: "wl-scale"}, {ID: "net-svc"}}})

	all, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	var ids []string
	for _, task := range all {
		ids = append(ids, task.ID)
	}
	if !slices.Equal(ids, []string{"net-svc", "wl-scale"}) {
		t.Errorf("ids = %q, want sorted", ids)
	}
}

func TestListTasksLeavesTheRepositorysSliceAlone(t *testing.T) {
	stored := []tasks.Task{{ID: "wl-scale"}, {ID: "net-svc"}}

	NewListTasks(fakeRepository{tasks: stored}).Execute(context.Background())

	if stored[0].ID != "wl-scale" {
		t.Error("sorting reordered the repository's own slice")
	}
}

func TestListTasksIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	uc := NewListTasks(fakeRepository{err: errors.New("disk gone")})

	_, err := uc.Execute(context.Background())

	if !errors.Is(err, tasks.ErrUnavailable) || !strings.Contains(err.Error(), "disk gone") {
		t.Errorf("got error %v", err)
	}
}

type fakeRepository struct {
	tasks []tasks.Task
	err   error
}

func (r fakeRepository) List(context.Context) ([]tasks.Task, error) { return r.tasks, r.err }
