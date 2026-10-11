package desk

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/exam"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

var errAssistantDuringExam = errors.New("the assistant is off during an exam")

type Desk struct {
	Lab        *sandbox.Lab
	Exam       *exam.Session
	busy       sync.Mutex
	now        func() time.Time
	mu         sync.Mutex
	touched    time.Time
	assistants map[*assistantShell]struct{}
}

func (d *Desk) Lock()         { d.busy.Lock() }
func (d *Desk) TryLock() bool { return d.busy.TryLock() }
func (d *Desk) Unlock()       { d.busy.Unlock() }

type Opener func(accountID string) (*Desk, error)

type Desks struct {
	open      Opener
	now       func() time.Time
	mu        sync.Mutex
	byAccount map[string]*Desk
}

func New(open Opener, now func() time.Time) *Desks {
	return &Desks{open: open, now: now, byAccount: map[string]*Desk{}}
}

func (ds *Desks) For(accountID string) (*Desk, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	d, ok := ds.byAccount[accountID]
	if !ok {
		var err error
		if d, err = ds.open(accountID); err != nil {
			return nil, err
		}
		d.now = ds.now
		d.resumeExam()
		ds.byAccount[accountID] = d
	}
	d.touch(ds.now())
	return d, nil
}

func (ds *Desks) Watch(ctx context.Context, every, idle time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ds.EndIdle(ctx, idle)
		}
	}
}

func (ds *Desks) EndIdle(ctx context.Context, idle time.Duration) {
	for id, d := range ds.snapshot() {
		if !d.idle(ds.now(), idle) || !d.busy.TryLock() {
			continue
		}
		// A request may have touched the desk while the lock was being taken.
		if d.idle(ds.now(), idle) {
			if err := d.Lab.EndBecause(ctx, fmt.Sprintf("ended after %d minutes idle", int(idle.Minutes()))); err != nil {
				log.Printf("end idle lab of %s: %v", id, err)
			} else {
				log.Printf("ended the lab of %s after %s idle", id, idle)
			}
		}
		d.busy.Unlock()
	}
}

func (ds *Desks) EndAll(ctx context.Context) error {
	var errs []error
	for _, d := range ds.snapshot() {
		errs = append(errs, d.Lab.End(ctx))
	}
	return errors.Join(errs...)
}

func (ds *Desks) snapshot() map[string]*Desk {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	desks := make(map[string]*Desk, len(ds.byAccount))
	for id, d := range ds.byAccount {
		desks[id] = d
	}
	return desks
}

func (d *Desk) Open(ctx context.Context, command string, size labagent.Size) (labagent.Shell, error) {
	shell, err := d.Lab.Open(ctx, command, size)
	if err != nil {
		return nil, err
	}
	return typing{Shell: shell, desk: d}, nil
}

func (d *Desk) OpenAssistant(ctx context.Context, command string, size labagent.Size) (labagent.Shell, error) {
	shell, err := d.Open(ctx, command, size)
	if err != nil {
		return nil, err
	}
	a := &assistantShell{Shell: shell, desk: d}
	d.mu.Lock()
	if d.assistants == nil {
		d.assistants = map[*assistantShell]struct{}{}
	}
	d.assistants[a] = struct{}{}
	d.mu.Unlock()
	// Checked after registering, so an exam beginning now either ends this shell or is seen here.
	if d.examRunning() {
		a.Close()
		return nil, errAssistantDuringExam
	}
	return a, nil
}

func (d *Desk) EndAssistants() {
	d.mu.Lock()
	open := slices.Collect(maps.Keys(d.assistants))
	d.mu.Unlock()
	for _, a := range open {
		a.Close()
	}
}

type assistantShell struct {
	labagent.Shell
	desk *Desk
}

func (a *assistantShell) Close() error {
	a.desk.mu.Lock()
	delete(a.desk.assistants, a)
	a.desk.mu.Unlock()
	return a.Shell.Close()
}

type typing struct {
	labagent.Shell
	desk *Desk
}

func (t typing) Write(keys []byte) (int, error) {
	t.desk.touch(t.desk.now())
	return t.Shell.Write(keys)
}

func (d *Desk) touch(now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if now.After(d.touched) {
		d.touched = now
	}
}

// A running exam keeps its lab however long the candidate thinks.
func (d *Desk) idle(now time.Time, after time.Duration) bool {
	if d.Lab.Status().State != sandbox.Ready || d.examRunning() {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return now.Sub(d.touched) >= after
}

func (d *Desk) examRunning() bool {
	return d.Exam != nil && d.Exam.Running()
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
