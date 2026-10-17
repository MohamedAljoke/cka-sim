package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type fakeRunner struct {
	result      labs.ScriptResult
	err         error
	host        string
	script      string
	ctxErr      error
	hasDeadline bool
}

func (r *fakeRunner) RunScript(ctx context.Context, host, script string) (labs.ScriptResult, error) {
	r.host, r.script, r.ctxErr = host, script, ctx.Err()
	_, r.hasDeadline = ctx.Deadline()
	return r.result, r.err
}
