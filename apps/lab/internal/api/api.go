package api

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster/usecases"
)

type api struct {
	labv1connect.UnimplementedLabServiceHandler
	clusters usecases.UseCases
}

func New(token string, clusters usecases.UseCases) http.Handler {
	server := connect.NewServer(logFailures, rpcauth.RequireBearer(token))
	labv1connect.RegisterLabServiceHandler(server, &api{clusters: clusters})
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}

func logFailures(next connect.ServerFunc) connect.ServerFunc {
	return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
		err := next(ctx, spec, stream)
		if err != nil && !isRefusal(connect.CodeOf(err)) {
			log.Printf("%s: %v", spec.Procedure, err)
		}

		return err
	}
}

func isRefusal(code connect.Code) bool {
	return code == connect.CodeFailedPrecondition || code == connect.CodeUnauthenticated
}
