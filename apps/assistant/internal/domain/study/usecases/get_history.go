package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
)

type GetHistory struct {
	platform Platform
}

func NewGetHistory(platform Platform) *GetHistory {
	return &GetHistory{platform: platform}
}

// Execute answers the student's attempts and notes newest first, only those of taskIDs unless it's empty.
func (uc *GetHistory) Execute(ctx context.Context, student string, taskIDs []string) ([]study.Entry, error) {
	return uc.platform.History(ctx, student, taskIDs)
}
