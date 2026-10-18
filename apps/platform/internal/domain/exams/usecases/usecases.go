package usecases

import (
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
) UseCases {
	perOwner := &sessions{all: map[string]*session{}, repo: repo, tasks: tasks, now: now, rng: rng}

	return UseCases{
		Begin:   &BeginExam{ss: perOwner, list: tasks.List, lab: lab},
		Get:     &GetExam{ss: perOwner},
		Flag:    &FlagTask{ss: perOwner},
		End:     &EndExam{ss: perOwner},
		Discard: &DiscardExam{ss: perOwner},
		Resume:  &ResumeExam{ss: perOwner},
	}
}
