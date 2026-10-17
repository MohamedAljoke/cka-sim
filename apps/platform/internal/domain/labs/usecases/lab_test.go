package usecases

import (
	"context"
	"errors"
	"fmt"
	"io"
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

func TestAShellOpensAtTheTerminalsSize(t *testing.T) {
	p := fakeProvider{}
	size := labs.TerminalSize{Cols: 120, Rows: 40}

	shell, err := NewOpenShell(p).Execute(context.Background(), size)

	if err != nil || shell.(fakeShell).size != size {
		t.Errorf("got %+v, %v", shell, err)
	}
}

func TestAShellOnALabThatIsNotReadyStaysNotReady(t *testing.T) {
	p := fakeProvider{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}

	_, err := NewOpenShell(p).Execute(context.Background(), labs.TerminalSize{Cols: 80, Rows: 24})

	if !errors.Is(err, labs.ErrNotReady) || errors.Is(err, labs.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestAShellThatCannotOpenMakesTheLabUnavailableAndKeepsTheCause(t *testing.T) {
	p := fakeProvider{err: errors.New("dial tcp 127.0.0.1:7171: connection refused")}

	_, err := NewOpenShell(p).Execute(context.Background(), labs.TerminalSize{Cols: 80, Rows: 24})

	if !errors.Is(err, labs.ErrUnavailable) || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("got %v", err)
	}
}

type fakeProvider struct {
	lab labs.Lab
	err error
}

func (p fakeProvider) Get(context.Context) (labs.Lab, error)   { return p.lab, p.err }
func (p fakeProvider) Start(context.Context) (labs.Lab, error) { return p.lab, p.err }
func (p fakeProvider) Stop(context.Context) (labs.Lab, error)  { return p.lab, p.err }

func (p fakeProvider) OpenShell(_ context.Context, size labs.TerminalSize) (Shell, error) {
	if p.err != nil {
		return nil, p.err
	}
	return fakeShell{size: size}, nil
}

type fakeShell struct {
	size labs.TerminalSize
}

func (fakeShell) Read([]byte) (int, error)       { return 0, io.EOF }
func (fakeShell) Write(p []byte) (int, error)    { return len(p), nil }
func (fakeShell) Resize(labs.TerminalSize) error { return nil }
func (fakeShell) Close() error                   { return nil }
