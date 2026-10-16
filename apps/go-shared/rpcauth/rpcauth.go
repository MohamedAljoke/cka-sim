package rpcauth

import (
	"context"
	"crypto/subtle"

	"connectrpc.com/connect/v2"
)

func Bearer(token string) connect.ClientInterceptor {
	return func(next connect.ClientFunc) connect.ClientFunc {
		return func(ctx context.Context, spec connect.Spec) (connect.ClientStream, error) {
			info, _ := connect.CallInfoForClientContext(ctx)
			info.RequestHeader().Set("Authorization", "Bearer "+token)

			return next(ctx, spec)
		}
	}
}

func RequireBearer(token string) connect.ServerInterceptor {
	want := []byte("Bearer " + token)

	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			info, _ := connect.CallInfoForServerContext(ctx)
			if subtle.ConstantTimeCompare([]byte(info.RequestHeader().Get("Authorization")), want) != 1 {
				return connect.NewError(connect.CodeUnauthenticated, "wrong or missing token")
			}

			return next(ctx, spec, stream)
		}
	}
}
