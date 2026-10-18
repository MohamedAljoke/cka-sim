package usecases

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	taskusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

// sessions keeps one exam per owner, read from the repository the first time the owner needs it.
type sessions struct {
	mu    sync.Mutex
	all   map[string]*session
	repo  ExamRepository
	tasks taskusecases.UseCases
	now   func() time.Time
	rng   *rand.Rand
}

func (ss *sessions) of(owner string) (*session, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if s := ss.all[owner]; s != nil {
		return s, nil
	}

	saved, err := ss.repo.Load(owner)
	if err != nil {
		return nil, fmt.Errorf("%w: read %s's exam: %w", exams.ErrUnavailable, owner, err)
	}

	s := &session{
		owner: owner, exam: saved, repo: ss.repo, now: ss.now,
		heal: ss.tasks.Heal, setUpTask: ss.tasks.SetUp, check: ss.tasks.Check, tidy: ss.tasks.Tidy,
		// A rand.Rand isn't safe to share, and each session draws under its own lock.
		rng: rand.New(rand.NewPCG(ss.rng.Uint64(), ss.rng.Uint64())), //nolint:gosec // drawing tasks, not secrets
	}
	ss.all[owner] = s

	return s, nil
}
