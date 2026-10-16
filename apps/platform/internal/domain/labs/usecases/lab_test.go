package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

func TestEachUseCaseReturnsWhatTheProviderSays(t *testing.T) {
	p := fakeProvider{lab: labs.Lab{State: labs.Starting, Step: "Preparing nodes"}}

	for name, execute := range map[string]func(context.Context) (labs.Lab, error){
		"get":   NewGetLab(p).Execute,
		"start": NewStartLab(p).Execute,
		"stop":  NewStopLab(p).Execute,
	} {
		got, err := execute(context.Background())

		if err != nil || got != p.lab {
			t.Errorf("%s: got %+v, %v", name, got, err)
		}
	}
}

func TestAProviderFailureMakesTheLabUnavailableAndKeepsTheCause(t *testing.T) {
	p := fakeProvider{err: errors.New("dial tcp 127.0.0.1:7171: connection refused")}

	for name, execute := range map[string]func(context.Context) (labs.Lab, error){
		"get":   NewGetLab(p).Execute,
		"start": NewStartLab(p).Execute,
		"stop":  NewStopLab(p).Execute,
	} {
		_, err := execute(context.Background())

		if !errors.Is(err, labs.ErrUnavailable) || !strings.Contains(err.Error(), "connection refused") {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

func TestABusyLabStaysBusy(t *testing.T) {
	p := fakeProvider{err: fmt.Errorf("%w: delete while creating", labs.ErrBusy)}

	for name, execute := range map[string]func(context.Context) (labs.Lab, error){
		"start": NewStartLab(p).Execute,
		"stop":  NewStopLab(p).Execute,
	} {
		_, err := execute(context.Background())

		if !errors.Is(err, labs.ErrBusy) || errors.Is(err, labs.ErrUnavailable) {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

type fakeProvider struct {
	lab labs.Lab
	err error
}

func (p fakeProvider) Get(context.Context) (labs.Lab, error)   { return p.lab, p.err }
func (p fakeProvider) Start(context.Context) (labs.Lab, error) { return p.lab, p.err }
func (p fakeProvider) Stop(context.Context) (labs.Lab, error)  { return p.lab, p.err }
