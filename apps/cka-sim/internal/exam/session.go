package exam

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/grader"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

var (
	ErrNoExam = errors.New("no exam")
	ErrExists = errors.New("an exam is already in progress")
)

type Config struct {
	Store   Store
	Files   fs.FS
	Runner  tasks.Runner
	Catalog []tasks.Task
	Now     func() time.Time
	Rand    *rand.Rand
}

// Session owns the one exam there can be. Every change is saved before it's visible.
type Session struct {
	cfg  Config
	mu   sync.Mutex
	exam *Exam
}

func Open(cfg Config) (*Session, error) {
	e, err := cfg.Store.Load()
	if err != nil {
		return nil, fmt.Errorf("load the saved exam: %w", err)
	}
	if e != nil {
		for _, t := range e.Tasks {
			if _, ok := tasks.Find(cfg.Catalog, t.ID); !ok {
				return nil, fmt.Errorf("the saved exam has task %q, which the catalog no longer has", t.ID)
			}
		}
	}
	return &Session{cfg: cfg, exam: e}, nil
}

func (s *Session) Current() (Exam, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.exam == nil {
		return Exam{}, false
	}
	return clone(*s.exam), true
}

func (s *Session) Running() bool {
	e, ok := s.Current()
	return ok && !e.Over()
}

func (s *Session) Weights() map[string]int {
	w := map[string]int{}
	for _, t := range s.cfg.Catalog {
		w[t.ID] = t.Weight
	}
	return w
}

// Begin draws and saves a new exam, then sets its tasks up (see setUp). Once all setups are
// done it calls release and starts the clock in one step, so whoever sees the clock running
// also sees what release freed.
func (s *Session) Begin(n, minutes int, release func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.exam != nil {
		return ErrExists
	}
	drawn := tasks.Draw(s.cfg.Catalog, n, s.cfg.Rand)
	if len(drawn) == 0 {
		return errors.New("the catalog has no tasks")
	}
	e := &Exam{Minutes: minutes}
	for _, t := range drawn {
		e.Tasks = append(e.Tasks, Task{ID: t.ID, Setup: Preparing})
	}
	if err := s.cfg.Store.Save(*e); err != nil {
		return err
	}
	s.exam = e
	s.setUp(e, release)
	return nil
}

// Resume finishes what a restart interrupted: setups that never ended, or a clock never started.
// It reports whether it had work to do; only then is release called, as in Begin.
func (s *Session) Resume(release func()) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.exam == nil || s.exam.Prepared() {
		return false
	}
	s.setUp(s.exam, release)
	return true
}

func (s *Session) Flag(id string, on bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.exam == nil {
		return ErrNoExam
	}
	if s.exam.Over() {
		return errors.New("the exam has ended")
	}
	i := slices.IndexFunc(s.exam.Tasks, func(t Task) bool { return t.ID == id })
	if i < 0 {
		return fmt.Errorf("task %q is not in this exam", id)
	}
	s.exam.Tasks[i].Flagged = on
	return s.cfg.Store.Save(*s.exam)
}

// End runs every check.sh in parallel and records the results; the exam is then over.
func (s *Session) End(ctx context.Context) (Exam, error) {
	s.mu.Lock()
	e := s.exam
	var err error
	switch {
	case e == nil:
		err = ErrNoExam
	case e.Over():
		err = errors.New("the exam has already ended")
	case !e.Prepared():
		err = errors.New("the exam is still being set up")
	}
	var ids []string
	if err == nil {
		for _, t := range e.Tasks {
			ids = append(ids, t.ID)
		}
	}
	s.mu.Unlock()
	if err != nil {
		return Exam{}, err
	}

	results := make([]grader.Result, len(ids))
	checkErrs := make([]error, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Go(func() {
			results[i], checkErrs[i] = tasks.Check(ctx, s.cfg.Files, s.cfg.Runner, s.task(id))
		})
	}
	wg.Wait()
	s.heal(ctx, func(t tasks.Task) bool { return slices.Contains(ids, t.ID) })

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.exam != e {
		return Exam{}, ErrNoExam
	}
	for i := range e.Tasks {
		if checkErrs[i] != nil {
			results[i] = grader.Result{Checks: []grader.Check{}}
			e.Tasks[i].CheckError = checkErrs[i].Error()
		}
		e.Tasks[i].Result = &results[i]
	}
	e.Ended = s.cfg.Now()
	if err := s.cfg.Store.Save(*e); err != nil {
		return Exam{}, err
	}
	return clone(*e), nil
}

func (s *Session) Discard() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.exam = nil
	return s.cfg.Store.Clear()
}

// setUp runs with s.mu held; its goroutines take the lock again to record each result.
// It first heals what earlier tasks left broken, except tasks this exam has already set up.
// Tasks marked last run one by one after the rest: an etcd snapshot taken in setup would lose
// objects other setups create later, and a broken scheduler or kubelet can stall setups beside it.
func (s *Session) setUp(e *Exam, release func()) {
	var done, together, last []string
	for _, t := range e.Tasks {
		if t.Setup != Preparing {
			done = append(done, t.ID)
		}
		switch {
		case t.Setup != Preparing:
		case s.task(t.ID).Order == tasks.Last:
			last = append(last, t.ID)
		default:
			together = append(together, t.ID)
		}
	}
	go func() {
		s.heal(context.Background(), func(t tasks.Task) bool { return !slices.Contains(done, t.ID) })
		var wg sync.WaitGroup
		for _, id := range together {
			wg.Go(func() { s.start(e, id) })
		}
		wg.Wait()
		for _, id := range last {
			s.start(e, id)
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		release()
		if s.exam == e {
			e.Started = s.cfg.Now()
			e.Deadline = e.Started.Add(time.Duration(e.Minutes) * time.Minute)
			s.saveInBackground(e)
		}
	}()
}

func (s *Session) start(e *Exam, id string) {
	err := tasks.Start(context.Background(), s.cfg.Files, s.cfg.Runner, s.task(id))
	s.update(e, func() {
		i := slices.IndexFunc(e.Tasks, func(t Task) bool { return t.ID == id })
		e.Tasks[i].Setup = Ready
		if err != nil {
			e.Tasks[i].Setup, e.Tasks[i].SetupError = Failed, err.Error()
		}
	})
}

// update changes e and saves it, unless the exam was discarded meanwhile.
func (s *Session) update(e *Exam, change func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.exam != e {
		return
	}
	change()
	s.saveInBackground(e)
}

// saveInBackground is for goroutines with no caller to return an error to.
func (s *Session) saveInBackground(e *Exam) {
	if err := s.cfg.Store.Save(*e); err != nil {
		log.Printf("save the exam: %v", err)
	}
}

// heal logs a failure instead of returning it: each task's own setup or check shows the damage.
func (s *Session) heal(ctx context.Context, keep func(tasks.Task) bool) {
	var some []tasks.Task
	for _, t := range s.cfg.Catalog {
		if keep(t) {
			some = append(some, t)
		}
	}
	if err := tasks.Heal(ctx, s.cfg.Files, s.cfg.Runner, some); err != nil {
		log.Printf("heal the cluster: %v", err)
	}
}

func (s *Session) task(id string) tasks.Task {
	t, _ := tasks.Find(s.cfg.Catalog, id)
	return t
}

func clone(e Exam) Exam {
	e.Tasks = slices.Clone(e.Tasks)
	return e
}
