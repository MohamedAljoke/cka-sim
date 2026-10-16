package rpcauth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestTheRightTokenIsLetThrough(t *testing.T) {
	url := serve(t, "s3cret")

	_, err := client(url, rpcauth.Bearer("s3cret")).GetCluster(context.Background(), &labv1.GetClusterRequest{})

	if err != nil {
		t.Fatalf("got %v", err)
	}
}

func TestAWrongOrMissingTokenIsUnauthenticated(t *testing.T) {
	url := serve(t, "s3cret")

	for name, c := range map[string]labv1connect.LabServiceClient{
		"wrong":   client(url, rpcauth.Bearer("guess")),
		"missing": client(url),
	} {
		_, err := c.GetCluster(context.Background(), &labv1.GetClusterRequest{})

		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("%s token: got %v", name, err)
		}
	}
}

func TestAnEmptyTokenNeverMatchesAMissingHeader(t *testing.T) {
	url := serve(t, "")

	_, err := client(url).GetCluster(context.Background(), &labv1.GetClusterRequest{})

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

type labService struct {
	labv1connect.UnimplementedLabServiceHandler
}

func (labService) GetCluster(context.Context, *labv1.GetClusterRequest) (*labv1.GetClusterResponse, error) {
	return &labv1.GetClusterResponse{}, nil
}

func serve(t *testing.T, token string) string {
	t.Helper()
	server := connect.NewServer(rpcauth.RequireBearer(token))
	labv1connect.RegisterLabServiceHandler(server, labService{})
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

func client(url string, interceptors ...connect.ClientInterceptor) labv1connect.LabServiceClient {
	return labv1connect.NewLabServiceClient(connect.NewClient(connecthttp.NewTransport(http.DefaultClient, url), interceptors...))
}
