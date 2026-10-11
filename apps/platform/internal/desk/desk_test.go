package desk

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

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
	})

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

func newDesks() (*Desks, *fakeProvider) {
	p := &fakeProvider{}
	return New(func(id string) (*Desk, error) {
		return &Desk{Lab: sandbox.NewLab(p, id, time.Now)}, nil
	}), p
}

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
	return sandbox.Box{ID: owner}, nil
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
