package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type fakeRunner struct {
	result      labs.ScriptResult
	results     map[string]labs.ScriptResult
	err         error
	host        string
	script      string
	ran         []string
	ctxErr      error
	hasDeadline bool
}

func (r *fakeRunner) RunScript(ctx context.Context, host, script string) (labs.ScriptResult, error) {
	r.host, r.script, r.ctxErr = host, script, ctx.Err()
	r.ran = append(r.ran, script)
	_, r.hasDeadline = ctx.Deadline()
	if result, ok := r.results[script]; ok {
		return result, r.err
	}
	return r.result, r.err
}
