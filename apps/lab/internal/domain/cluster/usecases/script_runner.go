package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

type ScriptRunner interface {
	Run(ctx context.Context, host, script string) (cluster.ScriptResult, error)
}
