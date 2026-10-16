package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestGetTaskGivesTheTaskWithItsQuestion(t *testing.T) {
	uc := NewGetTask(fakeRepository{tasks: []tasks.Task{{ID: "wl-scale", Question: "Scale it to 4."}}})

	got, err := uc.Execute(context.Background(), "wl-scale")

	if err != nil || got.ID != "wl-scale" || got.Question != "Scale it to 4." {
		t.Errorf("Execute = %+v, %v", got, err)
	}
}

func TestGetTaskIsNotFoundForAnUnknownID(t *testing.T) {
	uc := NewGetTask(fakeRepository{tasks: []tasks.Task{{ID: "wl-scale"}}})

	_, err := uc.Execute(context.Background(), "net-svc")

	if !errors.Is(err, tasks.ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestGetTaskIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	uc := NewGetTask(fakeRepository{err: errors.New("disk gone")})

	_, err := uc.Execute(context.Background(), "wl-scale")

	if !errors.Is(err, tasks.ErrUnavailable) || !strings.Contains(err.Error(), "disk gone") {
		t.Errorf("got error %v", err)
	}
}
