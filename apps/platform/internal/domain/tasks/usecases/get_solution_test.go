package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestGetSolutionGivesTheTasksSolution(t *testing.T) {
	uc := NewGetSolution(fakeRepository{tasks: []tasks.Task{{ID: "wl-scale", Solution: "kubectl scale."}}})

	got, err := uc.Execute(context.Background(), "wl-scale")

	if err != nil || got != "kubectl scale." {
		t.Errorf("Execute = %q, %v", got, err)
	}
}

func TestGetSolutionIsNotFoundForAnUnknownID(t *testing.T) {
	uc := NewGetSolution(fakeRepository{tasks: []tasks.Task{{ID: "wl-scale"}}})

	_, err := uc.Execute(context.Background(), "net-svc")

	if !errors.Is(err, tasks.ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestGetSolutionIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	uc := NewGetSolution(fakeRepository{err: errors.New("disk gone")})

	_, err := uc.Execute(context.Background(), "wl-scale")

	if !errors.Is(err, tasks.ErrUnavailable) || !strings.Contains(err.Error(), "disk gone") {
		t.Errorf("got error %v", err)
	}
}
