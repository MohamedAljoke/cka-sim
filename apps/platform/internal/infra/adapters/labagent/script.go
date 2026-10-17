package labagent

import (
	"context"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

func (p *Provider) RunScript(ctx context.Context, host, script string) (labs.ScriptResult, error) {
	res, err := p.work.RunScript(ctx, &labv1.RunScriptRequest{Host: host, Script: script})
	if err != nil {
		return labs.ScriptResult{}, notReady(err)
	}

	return labs.ScriptResult{Output: string(res.GetOutput()), ExitCode: int(res.GetExitCode())}, nil
}
