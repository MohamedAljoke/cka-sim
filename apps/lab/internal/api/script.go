package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

func (a *api) RunScript(ctx context.Context, req *labv1.RunScriptRequest) (*labv1.RunScriptResponse, error) {
	result, err := a.clusters.RunScript.Execute(ctx, req.GetHost(), req.GetScript())
	switch {
	case errors.Is(err, cluster.ErrNotReady):
		return nil, connect.NewError(connect.CodeFailedPrecondition, cluster.ErrNotReady.Message).WithCause(err)
	case errors.Is(err, cluster.ErrUnknownHost):
		return nil, connect.NewError(connect.CodeInvalidArgument, cluster.ErrUnknownHost.Message).WithCause(err)
	case err != nil:
		return nil, err
	}

	return &labv1.RunScriptResponse{Output: result.Output, ExitCode: int32(result.ExitCode)}, nil
}
