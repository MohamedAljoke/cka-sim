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

	all, err := uc.Execute(context.Background(), tasks.Filter{})
	if err != nil {
		t.Fatal(err)
	}

	if got := ids(all); !slices.Equal(got, []string{"net-svc", "wl-scale"}) {
		t.Errorf("ids = %q, want sorted", got)
	}
}

func TestListTasksLeavesTheRepositorysSliceAlone(t *testing.T) {
	stored := []tasks.Task{{ID: "wl-scale"}, {ID: "net-svc"}}

	NewListTasks(fakeRepository{tasks: stored}).Execute(context.Background(), tasks.Filter{})

	if stored[0].ID != "wl-scale" {
		t.Error("sorting reordered the repository's own slice")
	}
}

func TestListTasksKeepsOnlyWhatTheFilterMatches(t *testing.T) {
	uc := NewListTasks(fakeRepository{tasks: []tasks.Task{
		{ID: "nw-netpol", Domain: tasks.Networking, Topics: []string{"networkpolicy"}},
		{ID: "nw-dns", Domain: tasks.Networking, Topics: []string{"dns"}},
		{ID: "st-netpol", Domain: tasks.Storage, Topics: []string{"networkpolicy"}},
	}})

	got, err := uc.Execute(context.Background(), tasks.Filter{Domains: []tasks.Domain{tasks.Networking}, Topics: []string{"networkpolicy"}})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(ids(got), []string{"nw-netpol"}) {
		t.Errorf("ids = %q", ids(got))
	}
}

func TestListTasksMatchingNothingIsEmptyNotNil(t *testing.T) {
	got, err := NewListTasks(fakeRepository{tasks: []tasks.Task{{ID: "wl-scale", Domain: tasks.Workloads}}}).
		Execute(context.Background(), tasks.Filter{Domains: []tasks.Domain{tasks.Storage}})

	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("got %v, %v; want an empty list", got, err)
	}
}

func TestListTasksRefusesAnInvalidFilterBeforeLoading(t *testing.T) {
	uc := NewListTasks(fakeRepository{err: errors.New("disk gone")})

	_, err := uc.Execute(context.Background(), tasks.Filter{Domains: []tasks.Domain{"security"}})

	if !errors.Is(err, tasks.ErrUnknownDomain) || errors.Is(err, tasks.ErrUnavailable) {
		t.Errorf("got error %v, want only ErrUnknownDomain", err)
	}
}

func TestListTasksIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	uc := NewListTasks(fakeRepository{err: errors.New("disk gone")})

	_, err := uc.Execute(context.Background(), tasks.Filter{})

	if !errors.Is(err, tasks.ErrUnavailable) || !strings.Contains(err.Error(), "disk gone") {
		t.Errorf("got error %v", err)
	}
}

func ids(all []tasks.Task) []string {
	var ids []string
	for _, task := range all {
		ids = append(ids, task.ID)
	}
	return ids
}
