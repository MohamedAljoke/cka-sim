package usecases

import (
	"fmt"
	"math/rand/v2"
	"time"

	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	taskusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

type UseCases struct {
	Begin   *BeginExam
	Get     *GetExam
	Flag    *FlagTask
	End     *EndExam
	Discard *DiscardExam
	Resume  *ResumeExam
}

func New(
	repo ExamRepository, tasks taskusecases.UseCases, lab *labusecases.GetLab, now func() time.Time, rng *rand.Rand,
) (UseCases, error) {
	saved, err := repo.Load()
	if err != nil {
		return UseCases{}, fmt.Errorf("load the saved exam: %w", err)
	}

	s := &session{exam: saved, repo: repo, heal: tasks.Heal, setUpTask: tasks.SetUp, check: tasks.Check,
		tidy: tasks.Tidy, now: now, rng: rng}

	return UseCases{
		Begin:   &BeginExam{s: s, list: tasks.List, lab: lab},
		Get:     &GetExam{s: s},
		Flag:    &FlagTask{s: s},
		End:     &EndExam{s: s},
		Discard: &DiscardExam{s: s},
		Resume:  &ResumeExam{s: s},
	}, nil
}
