package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

type InspectCluster struct {
	platform Platform
}

func NewInspectCluster(platform Platform) *InspectCluster {
	return &InspectCluster{platform: platform}
}

// Execute runs kubectl with args on the student's cluster; the platform decides which commands are read-only.
func (uc *InspectCluster) Execute(ctx context.Context, student string, args []string) (tutor.Inspection, error) {
	return uc.platform.Inspect(ctx, student, args)
}
