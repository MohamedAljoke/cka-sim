package sandbox

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/terminal"
)

func TestNewLabHasNoBox(t *testing.T) {
	lab := NewLab(&fakeProvider{}, fakeClock())

	if got := lab.Status(); got.State != None || got.Provider != "Fake" {
		t.Fatalf("status = %+v, want none from Fake", got)
	}
}

func TestStartMakesTheLabReady(t *testing.T) {
	p := &fakeProvider{gate: make(chan struct{})}
	lab := NewLab(p, fakeClock())

	lab.Start()
	if got := lab.Status().State; got != Starting {
		t.Fatalf("state = %s while creating, want %s", got, Starting)
	}
	close(p.gate)
	ensure(t, lab)

	got := lab.Status()
	if got.State != Ready || got.Error != "" {
		t.Fatalf("status = %+v, want ready", got)
	}
	if !got.Ready.After(got.Started) {
		t.Fatalf("ready %v is not after started %v", got.Ready, got.Started)
	}
}

func TestRunAndOpenNeedAReadyLab(t *testing.T) {
	lab := NewLab(&fakeProvider{}, fakeClock())

	if _, err := lab.Run(context.Background(), "host", "task", nil); !errors.Is(err, ErrNoLab) {
		t.Fatalf("Run before start = %v, want ErrNoLab", err)
	}
	if _, err := lab.Open(context.Background(), terminal.Size{}); !errors.Is(err, ErrNoLab) {
		t.Fatalf("Open before start = %v, want ErrNoLab", err)
	}
}

func TestRunAndOpenReachTheBox(t *testing.T) {
	p := &fakeProvider{}
	lab := NewLab(p, fakeClock())
	lab.Start()
	ensure(t, lab)

	out, err := lab.Run(context.Background(), "host", "task", []byte("echo"))
	if err != nil || out != "ran task on host" {
		t.Fatalf("Run = %q, %v", out, err)
	}
	if _, err := lab.Open(context.Background(), terminal.Size{}); err != nil {
		t.Fatalf("Open = %v", err)
	}
	if p.shells.opened != 1 {
		t.Fatalf("opened %d shells, want 1", p.shells.opened)
	}
}

func TestEnsureStartsAndWaits(t *testing.T) {
	p := &fakeProvider{gate: make(chan struct{})}
	lab := NewLab(p, fakeClock())

	done := make(chan error)
	go func() { done <- lab.Ensure(context.Background()) }()
	select {
	case err := <-done:
		t.Fatalf("Ensure returned %v before the box was created", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(p.gate)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestEnsureStopsWithItsContext(t *testing.T) {
	lab := NewLab(&fakeProvider{gate: make(chan struct{})}, fakeClock())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := lab.Ensure(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Ensure = %v, want context.Canceled", err)
	}
}

func TestAFailedStartCanBeRetried(t *testing.T) {
	p := &fakeProvider{err: errors.New("cluster is not up")}
	lab := NewLab(p, fakeClock())

	if err := lab.Ensure(context.Background()); err == nil {
		t.Fatal("Ensure succeeded, want the create error")
	}
	if got := lab.Status(); got.State != Failed || got.Error != "cluster is not up" {
		t.Fatalf("status = %+v, want failed with the error", got)
	}
	p.setErr(nil)
	ensure(t, lab)
	if got := lab.Status(); got.State != Ready || got.Error != "" {
		t.Fatalf("status after retry = %+v, want ready", got)
	}
}

func TestStartingTwiceCreatesOneBox(t *testing.T) {
	p := &fakeProvider{gate: make(chan struct{})}
	lab := NewLab(p, fakeClock())

	lab.Start()
	lab.Start()
	close(p.gate)
	ensure(t, lab)
	lab.Start()

	if got := p.count(); got != 1 {
		t.Fatalf("created %d boxes, want 1", got)
	}
}

func TestEndDestroysTheBox(t *testing.T) {
	p := &fakeProvider{}
	lab := NewLab(p, fakeClock())
	lab.Start()
	ensure(t, lab)

	if err := lab.End(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.destroyed != 1 {
		t.Fatalf("destroyed %d boxes, want 1", p.destroyed)
	}
	if got := lab.Status(); got.State != None {
		t.Fatalf("state = %s, want %s", got.State, None)
	}
	if _, err := lab.Run(context.Background(), "host", "task", nil); !errors.Is(err, ErrNoLab) {
		t.Fatalf("Run after end = %v, want ErrNoLab", err)
	}
}

func TestEndWithoutABoxDoesNothing(t *testing.T) {
	p := &fakeProvider{}
	lab := NewLab(p, fakeClock())

	if err := lab.End(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.destroyed != 0 {
		t.Fatalf("destroyed %d boxes, want 0", p.destroyed)
	}
}

func ensure(t *testing.T, lab *Lab) {
	t.Helper()
	if err := lab.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// fakeClock moves a second forward on every reading.
func fakeClock() func() time.Time {
	var mu sync.Mutex
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		now = now.Add(time.Second)
		return now
	}
}

// fakeProvider's Create waits for gate when it is set.
type fakeProvider struct {
	gate      chan struct{}
	mu        sync.Mutex
	err       error
	created   int
	destroyed int
	shells    fakeShells
}

func (p *fakeProvider) Name() string { return "Fake" }

func (p *fakeProvider) Create(ctx context.Context) (Box, error) {
	if p.gate != nil {
		<-p.gate
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.created++
	if p.err != nil {
		return Box{}, p.err
	}
	return Box{ID: "fake", Runner: fakeRunner{}, Shells: &p.shells}, nil
}

func (p *fakeProvider) Destroy(ctx context.Context, b Box) error {
	p.destroyed++
	return nil
}

func (p *fakeProvider) setErr(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.err = err
}

func (p *fakeProvider) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.created
}

type fakeRunner struct{}

func (fakeRunner) Run(ctx context.Context, host, taskID string, script []byte) (string, error) {
	return "ran " + taskID + " on " + host, nil
}

type fakeShells struct {
	opened int
}

func (s *fakeShells) Open(ctx context.Context, size terminal.Size) (terminal.Shell, error) {
	s.opened++
	return nil, nil
}

func (s *fakeShells) EndAll(ctx context.Context) error { return nil }
