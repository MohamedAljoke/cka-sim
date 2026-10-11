package desk

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

func TestForGivesAnAccountTheSameDeskEveryTime(t *testing.T) {
	desks, _ := newDesks()

	first, _ := desks.For("a")
	again, _ := desks.For("a")

	if first != again {
		t.Error("the same account got two desks")
	}
}

func TestForGivesEachAccountItsOwnLab(t *testing.T) {
	desks, p := newDesks()

	a, _ := desks.For("a")
	b, _ := desks.For("b")
	if err := a.Lab.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}

	if a == b || a.Lab == b.Lab {
		t.Fatal("two accounts share a desk")
	}
	if b.Lab.Status().State != sandbox.None {
		t.Errorf("b's lab is %s, want none", b.Lab.Status().State)
	}
	if got := p.owners(); len(got) != 1 || got[0] != "a" {
		t.Errorf("created labs for %q, want just a", got)
	}
}

func TestForReportsAFailedOpenAndTriesAgainNextTime(t *testing.T) {
	calls := 0
	desks := New(func(string) (*Desk, error) {
		calls++
		return nil, errors.New("bad exam file")
	}, time.Now)

	_, first := desks.For("a")
	_, second := desks.For("a")

	if first == nil || second == nil || calls != 2 {
		t.Errorf("errors %v and %v after %d opens, want two failures", first, second, calls)
	}
}

func TestEndAllEndsEveryLab(t *testing.T) {
	desks, p := newDesks()
	for _, id := range []string{"a", "b"} {
		d, _ := desks.For(id)
		if err := d.Lab.Ensure(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	if err := desks.EndAll(context.Background()); err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"a", "b"} {
		if d, _ := desks.For(id); d.Lab.Status().State != sandbox.None {
			t.Errorf("%s's lab is %s, want none", id, d.Lab.Status().State)
		}
	}
	if p.destroyed != 2 {
		t.Errorf("destroyed %d labs, want 2", p.destroyed)
	}
}

func TestEndIdleEndsOnlyLabsUntouchedForTheIdleTime(t *testing.T) {
	clock := &fakeClock{now: time.Now()}
	desks, p := newDesksAt(clock.Now)
	quiet := readyDesk(t, desks, "quiet")
	clock.advance(20 * time.Minute)
	busy := readyDesk(t, desks, "busy")

	clock.advance(10 * time.Minute)
	desks.EndIdle(context.Background(), 30*time.Minute)

	if quiet.Lab.Status().State != sandbox.None || busy.Lab.Status().State != sandbox.Ready {
		t.Errorf("quiet is %s and busy is %s, want none and ready", quiet.Lab.Status().State, busy.Lab.Status().State)
	}
	if p.destroyed != 1 {
		t.Errorf("destroyed %d labs, want 1", p.destroyed)
	}
	if got := quiet.Lab.Status().Ended; got != "ended after 30 minutes idle" {
		t.Errorf("ended = %q, want why", got)
	}
}

func TestTypingInTheTerminalKeepsTheLab(t *testing.T) {
	clock := &fakeClock{now: time.Now()}
	desks, _ := newDesksAt(clock.Now)
	d := readyDesk(t, desks, "a")
	shell, err := d.Open(context.Background(), labagent.Size{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}

	clock.advance(29 * time.Minute)
	shell.Write([]byte("k get pods\n"))
	clock.advance(29 * time.Minute)
	desks.EndIdle(context.Background(), 30*time.Minute)

	if d.Lab.Status().State != sandbox.Ready {
		t.Errorf("lab is %s after typing, want ready", d.Lab.Status().State)
	}
}

func TestEndIdleSkipsADeskRunningAScript(t *testing.T) {
	clock := &fakeClock{now: time.Now()}
	desks, _ := newDesksAt(clock.Now)
	d := readyDesk(t, desks, "a")
	d.Lock()
	defer d.Unlock()

	clock.advance(time.Hour)
	desks.EndIdle(context.Background(), 30*time.Minute)

	if d.Lab.Status().State != sandbox.Ready {
		t.Errorf("lab is %s while a script runs, want ready", d.Lab.Status().State)
	}
}

func newDesks() (*Desks, *fakeProvider) {
	return newDesksAt(time.Now)
}

func newDesksAt(now func() time.Time) (*Desks, *fakeProvider) {
	p := &fakeProvider{}
	return New(func(id string) (*Desk, error) {
		return &Desk{Lab: sandbox.NewLab(p, id, time.Now)}, nil
	}, now), p
}

func readyDesk(t *testing.T, desks *Desks, id string) *Desk {
	t.Helper()
	d, err := desks.For(id)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Lab.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
	return d
}

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time          { return c.now }
func (c *fakeClock) advance(d time.Duration) { c.now = c.now.Add(d) }

type fakeShell struct{}

func (fakeShell) Read([]byte) (int, error)                    { return 0, nil }
func (fakeShell) Write(b []byte) (int, error)                 { return len(b), nil }
func (fakeShell) Resize(context.Context, labagent.Size) error { return nil }
func (fakeShell) Close() error                                { return nil }

type fakeShells struct{}

func (fakeShells) Open(context.Context, labagent.Size) (labagent.Shell, error) {
	return fakeShell{}, nil
}
func (fakeShells) EndAll(context.Context) error { return nil }

type fakeProvider struct {
	mu        sync.Mutex
	created   []string
	destroyed int
}

func (*fakeProvider) Name() string { return "Fake" }

func (p *fakeProvider) Create(_ context.Context, owner string) (sandbox.Box, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.created = append(p.created, owner)
	return sandbox.Box{ID: owner, Shells: fakeShells{}}, nil
}

func (p *fakeProvider) Destroy(context.Context, sandbox.Box) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.destroyed++
	return nil
}

func (p *fakeProvider) owners() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.created...)
}
