package exam

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/grader"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

var (
	ErrNoExam    = domainerr.NotFound("no_exam", "no exam")
	ErrExists    = domainerr.Conflict("exam_exists", "an exam is already in progress")
	ErrScoring   = domainerr.Conflict("exam_scoring", "the last exam is still being scored")
	ErrNoMatch   = domainerr.Unprocessable("no_matching_tasks", "no task matches those domains and topics")
	ErrEnded     = domainerr.Conflict("exam_ended", "the exam has already ended")
	ErrSettingUp = domainerr.Conflict("exam_setting_up", "the exam is still being set up")
)

type Config struct {
	Store   Store
	Files   fs.FS
	Runner  tasks.Runner
	Catalog []tasks.Task
	Now     func() time.Time
	Rand    *rand.Rand
	Ready   func(ctx context.Context) error
}

// Every change is saved before it's visible.
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

// release is called once every setup is done.
func (s *Session) Begin(pool []tasks.Task, n, minutes int, release func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch {
	case s.exam == nil:
	case !s.exam.Over():
		return ErrExists
	case s.exam.Scoring():
		return ErrScoring
	}
	drawn := tasks.Draw(pool, n, s.cfg.Rand)
	if len(drawn) == 0 {
		return ErrNoMatch
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

// Only when Resume reports work to do is release called.
func (s *Session) Resume(release func()) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch {
	case s.exam == nil:
		return false
	case !s.exam.Prepared() || s.exam.SettingUp():
		s.setUp(s.exam, release)
	case s.exam.Scoring():
		s.grade(s.exam, release)
	default:
		return false
	}
	return true
}

func (s *Session) Flag(id string, on bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.exam == nil {
		return ErrNoExam
	}
	if s.exam.Over() {
		return ErrEnded
	}
	i := slices.IndexFunc(s.exam.Tasks, func(t Task) bool { return t.ID == id })
	if i < 0 {
		return domainerr.NotFound("task_not_in_exam", fmt.Sprintf("task %q is not in this exam", id))
	}
	s.exam.Tasks[i].Flagged = on
	return s.cfg.Store.Save(*s.exam)
}

// Grading runs in the background; release is called when it's done.
func (s *Session) End(release func()) (Exam, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.exam
	switch {
	case e == nil:
		return Exam{}, ErrNoExam
	case e.Over():
		return Exam{}, ErrEnded
	case !e.Prepared() || e.SettingUp():
		return Exam{}, ErrSettingUp
	}
	e.Ended = s.cfg.Now()
	if err := s.cfg.Store.Save(*e); err != nil {
		e.Ended = time.Time{}
		return Exam{}, err
	}
	s.grade(e, release)
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
// Tasks marked last run after the rest: an etcd snapshot taken in setup would lose objects other
// setups create later, and a broken scheduler or kubelet can stall setups beside it. They run one
// by one per host, and hosts side by side: a worker's broken kubelet doesn't touch the control plane.
// The clock starts before them, so the user reads while they run, and their time is given back.
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
		err := s.ready()
		if err != nil {
			s.fail(e, append(together, last...), err)
		} else {
			s.heal(context.Background(), func(t tasks.Task) bool { return !slices.Contains(done, t.ID) })
			var wg sync.WaitGroup
			for _, id := range together {
				wg.Go(func() { s.start(e, id) })
			}
			wg.Wait()
		}
		s.update(e, func() {
			if !e.Prepared() {
				e.Started = s.cfg.Now()
				e.Deadline = e.Started.Add(time.Duration(e.Minutes) * time.Minute)
			}
		})
		began := s.cfg.Now()
		if err == nil {
			queues := map[string][]string{}
			for _, id := range last {
				host := s.task(id).Host
				queues[host] = append(queues[host], id)
			}
			var wg sync.WaitGroup
			for _, ids := range queues {
				wg.Go(func() {
					for _, id := range ids {
						s.start(e, id)
					}
				})
			}
			wg.Wait()
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		release()
		if s.exam == e && len(last) > 0 && err == nil {
			e.Deadline = e.Deadline.Add(s.cfg.Now().Sub(began))
			s.saveInBackground(e)
		}
	}()
}

// grade runs with s.mu held, like setUp. Every ungraded task's check.sh runs at once and each
// result is saved as it lands; a check that crashes scores 0 and keeps its error. Then the exam's
// tasks are tidied, so a kubelet left broken doesn't outlive the exam and the next setups find
// no namespace to delete.
func (s *Session) grade(e *Exam, release func()) {
	var ids []string
	for _, t := range e.Tasks {
		if t.Result == nil {
			ids = append(ids, t.ID)
		}
	}
	go func() {
		var wg sync.WaitGroup
		for _, id := range ids {
			wg.Go(func() { s.check(e, id) })
		}
		wg.Wait()
		s.tidy(e)
		s.mu.Lock()
		defer s.mu.Unlock()
		release()
		if s.exam == e {
			e.Scored = s.cfg.Now()
			s.saveInBackground(e)
		}
	}()
}

func (s *Session) check(e *Exam, id string) {
	began := time.Now()
	result, err := tasks.Check(context.Background(), s.cfg.Files, s.cfg.Runner, s.task(id))
	log.Printf("exam check %s: %s", id, since(began))
	s.update(e, func() {
		i := slices.IndexFunc(e.Tasks, func(t Task) bool { return t.ID == id })
		if err != nil {
			result = grader.Result{Checks: []grader.Check{}}
			e.Tasks[i].CheckError = err.Error()
		}
		e.Tasks[i].Result = &result
	})
}

func (s *Session) start(e *Exam, id string) {
	began := time.Now()
	err := tasks.Start(context.Background(), s.cfg.Files, s.cfg.Runner, s.task(id))
	log.Printf("exam setup %s: %s", id, since(began))
	s.update(e, func() {
		i := slices.IndexFunc(e.Tasks, func(t Task) bool { return t.ID == id })
		e.Tasks[i].Setup = Ready
		if err != nil {
			e.Tasks[i].Setup, e.Tasks[i].SetupError = Failed, err.Error()
		}
	})
}

func (s *Session) ready() error {
	if s.cfg.Ready == nil {
		return nil
	}
	return s.cfg.Ready(context.Background())
}

func (s *Session) fail(e *Exam, ids []string, err error) {
	s.update(e, func() {
		for i, t := range e.Tasks {
			if slices.Contains(ids, t.ID) {
				e.Tasks[i].Setup, e.Tasks[i].SetupError = Failed, err.Error()
			}
		}
	})
}

func (s *Session) update(e *Exam, change func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.exam != e {
		return
	}
	change()
	s.saveInBackground(e)
}

func (s *Session) saveInBackground(e *Exam) {
	if err := s.cfg.Store.Save(*e); err != nil {
		log.Printf("save the exam: %v", err)
	}
}

// Logged, not returned: each task's own setup or check shows the damage.
func (s *Session) heal(ctx context.Context, keep func(tasks.Task) bool) {
	var some []tasks.Task
	for _, t := range s.cfg.Catalog {
		if keep(t) {
			some = append(some, t)
		}
	}
	began := time.Now()
	if err := tasks.Heal(ctx, s.cfg.Files, s.cfg.Runner, some); err != nil {
		log.Printf("heal the cluster: %v", err)
	}
	log.Printf("exam heal: %s", since(began))
}

// Logged, not returned: the next setup's heal and fresh_ns are the safety net.
func (s *Session) tidy(e *Exam) {
	var some []tasks.Task
	for _, t := range e.Tasks {
		some = append(some, s.task(t.ID))
	}
	began := time.Now()
	if err := tasks.Tidy(context.Background(), s.cfg.Files, s.cfg.Runner, some); err != nil {
		log.Printf("tidy the cluster: %v", err)
	}
	log.Printf("exam tidy: %s", since(began))
}

func (s *Session) task(id string) tasks.Task {
	t, _ := tasks.Find(s.cfg.Catalog, id)
	return t
}

func since(t time.Time) time.Duration { return time.Since(t).Round(100 * time.Millisecond) }

func clone(e Exam) Exam {
	e.Tasks = slices.Clone(e.Tasks)
	return e
}
