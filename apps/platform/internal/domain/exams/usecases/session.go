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

// session is one owner's exam, which every use case shares. A request's change is saved before it shows; setups
// and checks run in the background and take the lock again to record each result.
type session struct {
	owner     string
	mu        sync.Mutex
	exam      *exams.Exam
	repo      ExamRepository
	heal      *taskusecases.HealCluster
	setUpTask *taskusecases.SetUpTask
	check     *taskusecases.CheckTask
	tidy      *taskusecases.TidyTask
	now       func() time.Time
	rng       *rand.Rand
	work      sync.WaitGroup
	// tidying holds the next exam's setups back until the last exam's tidy is done, so it can't delete the
	// namespaces they make.
	tidying sync.WaitGroup
}

func (s *session) save(e exams.Exam) error {
	if err := s.repo.Save(s.owner, e); err != nil {
		return fmt.Errorf("%w: %w", exams.ErrUnavailable, err)
	}

	return nil
}

// A background change stays in memory when the save fails, so the exam goes on; the next save catches the file up.
func (s *session) saveInBackground() {
	if err := s.repo.Save(s.owner, *s.exam); err != nil {
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

// Setups and checks outlive the request that started them, as a practice setup does. What earlier tasks broke is
// healed first; a heal that fails doesn't stop the exam, and the setups say if the cluster is broken.
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
		s.tidying.Wait()

		if healed, err := s.heal.Execute(ctx, s.owner); err != nil || !healed.Passed() {
			slog.Warn("heal the cluster before the exam", "err", err, "output", tail(healed.Output))
		}

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
	result, err := s.setUpTask.Execute(ctx, s.owner, id)

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

// Once scored, every task of the exam is tidied, so the next exam starts on a clean cluster.
func (s *session) grade(ctx context.Context) {
	ctx = context.WithoutCancel(ctx)

	var ids []string

	all := make([]string, 0, len(s.exam.Tasks))

	for _, task := range s.exam.Tasks {
		all = append(all, task.ID)
		if task.Setup != exams.SetupFailed && task.Grade == nil {
			ids = append(ids, task.ID)
		}
	}

	s.tidying.Add(1)

	s.work.Go(func() {
		defer s.tidying.Done()

		var wg sync.WaitGroup
		for _, id := range ids {
			wg.Go(func() { s.checkTask(ctx, id) })
		}

		wg.Wait()
		s.scored()

		if err := s.tidy.Execute(ctx, s.owner, all...); err != nil {
			slog.Error("tidy after the exam", "err", err)
		}
	})
}

func (s *session) scored() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.exam.Scored = s.now()
	s.saveInBackground()
}

// A check that breaks scores 0 and keeps why, so the results can say so.
func (s *session) checkTask(ctx context.Context, id string) {
	checked, err := s.check.Execute(ctx, s.owner, id)

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
