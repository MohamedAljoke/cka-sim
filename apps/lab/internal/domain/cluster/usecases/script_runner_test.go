package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

type fakeScriptRunner struct {
	result cluster.ScriptResult
	ran    bool
}

func (r *fakeScriptRunner) Run(context.Context, string, string) (cluster.ScriptResult, error) {
	r.ran = true
	return r.result, nil
}
