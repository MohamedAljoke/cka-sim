package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type ScriptRunner interface {
	RunScript(ctx context.Context, host, script string) (labs.ScriptResult, error)
}
