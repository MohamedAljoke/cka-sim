package desk

import (
	"context"
	"errors"
	"sync"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/exam"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

type Desk struct {
	Lab  *sandbox.Lab
	Exam *exam.Session
	busy sync.Mutex
}

func (d *Desk) Lock()         { d.busy.Lock() }
func (d *Desk) TryLock() bool { return d.busy.TryLock() }
func (d *Desk) Unlock()       { d.busy.Unlock() }

type Opener func(accountID string) (*Desk, error)

type Desks struct {
	open      Opener
	mu        sync.Mutex
	byAccount map[string]*Desk
}

func New(open Opener) *Desks {
	return &Desks{open: open, byAccount: map[string]*Desk{}}
}

func (ds *Desks) For(accountID string) (*Desk, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if d, ok := ds.byAccount[accountID]; ok {
		return d, nil
	}
	d, err := ds.open(accountID)
	if err != nil {
		return nil, err
	}
	d.resumeExam()
	ds.byAccount[accountID] = d
	return d, nil
}

func (ds *Desks) EndAll(ctx context.Context) error {
	ds.mu.Lock()
	desks := make([]*Desk, 0, len(ds.byAccount))
	for _, d := range ds.byAccount {
		desks = append(desks, d)
	}
	ds.mu.Unlock()
	var errs []error
	for _, d := range desks {
		errs = append(errs, d.Lab.End(ctx))
	}
	return errors.Join(errs...)
}

func (d *Desk) resumeExam() {
	if d.Exam == nil {
		return
	}
	d.busy.Lock()
	if !d.Exam.Resume(d.busy.Unlock) {
		d.busy.Unlock()
	}
}
