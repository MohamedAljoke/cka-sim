package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	taskusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

const outputTail = 2000

// session is the one exam every use case shares. A request's change is saved before it shows; setups and checks run
// in the background and take the lock again to record each result.
type session struct {
	mu    sync.Mutex
	exam  *exams.Exam
	repo  ExamRepository
	start *taskusecases.StartTask
	check *taskusecases.CheckTask
	now   func() time.Time
	rng   *rand.Rand
	work  sync.WaitGroup
}

func (s *session) save(e exams.Exam) error {
	if err := s.repo.Save(e); err != nil {
		return fmt.Errorf("%w: %w", exams.ErrUnavailable, err)
	}

	return nil
}

// A background change stays in memory when the save fails, so the exam goes on; the next save catches the file up.
func (s *session) saveInBackground() {
	if err := s.repo.Save(*s.exam); err != nil {
		slog.Error("save the exam", "err", err)
	}
}

func (s *session) endIfExpired(ctx context.Context) {
	if s.exam == nil || !s.exam.Expired(s.now()) {
		return
	}

	s.exam.Ended = s.exam.Deadline
	s.saveInBackground()
	s.grade(ctx)
}

// Setups and checks outlive the request that started them, as a practice setup does.
func (s *session) setUp(ctx context.Context) {
	ctx = context.WithoutCancel(ctx)

	var together, last []string

	for _, task := range s.exam.Tasks {
		switch {
		case task.Setup != exams.Preparing:
		case task.Last:
			last = append(last, task.ID)
		default:
			together = append(together, task.ID)
		}
	}

	s.work.Go(func() {
		var wg sync.WaitGroup
		for _, id := range together {
			wg.Go(func() { s.startTask(ctx, id) })
		}

		wg.Wait()

		for _, id := range last {
			s.startTask(ctx, id)
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		s.exam.Started = s.now()
		s.exam.Deadline = s.exam.Started.Add(time.Duration(s.exam.Minutes) * time.Minute)
		s.saveInBackground()
	})
}

func (s *session) startTask(ctx context.Context, id string) {
	result, err := s.start.Execute(ctx, id)

	s.mu.Lock()
	defer s.mu.Unlock()

	task := &s.exam.Tasks[s.exam.Index(id)]
	task.Setup = exams.SetUp

	switch {
	case err != nil:
		task.Setup, task.SetupError = exams.SetupFailed, err.Error()
	case !result.Passed():
		task.Setup = exams.SetupFailed
		task.SetupError = fmt.Sprintf("the setup exited with code %d\n%s", result.ExitCode, tail(result.Output))
	}

	s.saveInBackground()
}

func (s *session) grade(ctx context.Context) {
	ctx = context.WithoutCancel(ctx)

	var ids []string

	for _, task := range s.exam.Tasks {
		if task.Setup != exams.SetupFailed && task.Grade == nil {
			ids = append(ids, task.ID)
		}
	}

	s.work.Go(func() {
		var wg sync.WaitGroup
		for _, id := range ids {
			wg.Go(func() { s.checkTask(ctx, id) })
		}

		wg.Wait()

		s.mu.Lock()
		defer s.mu.Unlock()

		s.exam.Scored = s.now()
		s.saveInBackground()
	})
}

// A check that breaks scores 0 and keeps why, so the results can say so.
func (s *session) checkTask(ctx context.Context, id string) {
	checked, err := s.check.Execute(ctx, id)

	s.mu.Lock()
	defer s.mu.Unlock()

	task := &s.exam.Tasks[s.exam.Index(id)]
	task.Grade = &checked.Grade

	switch {
	case err != nil:
		task.Grade, task.CheckError = &tasks.Grade{Checks: []tasks.Check{}}, err.Error()
	case len(checked.Grade.Checks) == 0:
		task.CheckError = fmt.Sprintf("the check printed no result and exited with code %d\n%s",
			checked.Script.ExitCode, tail(checked.Script.Output))
	}

	s.saveInBackground()
}

func tail(output string) string {
	if len(output) <= outputTail {
		return output
	}

	return output[len(output)-outputTail:]
}
