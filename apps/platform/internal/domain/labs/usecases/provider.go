package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type LabProvider interface {
	Get(ctx context.Context) (labs.Lab, error)
	Start(ctx context.Context) (labs.Lab, error)
	Stop(ctx context.Context) (labs.Lab, error)
}
