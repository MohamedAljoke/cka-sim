package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
)

type ListTasks struct {
	platform Platform
}

func NewListTasks(platform Platform) *ListTasks {
	return &ListTasks{platform: platform}
}

func (uc *ListTasks) Execute(ctx context.Context, student string, filter study.Filter) ([]study.TaskSummary, error) {
	return uc.platform.ListTasks(ctx, student, filter)
}
