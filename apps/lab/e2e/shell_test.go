package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect/v2"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestAShellNeedsAReadyCluster(t *testing.T) {
	if *create {
		t.Skip("the -create run changes the cluster")
	}

	_, err := openShell(t, authed()).Receive()

	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("got %v", err)
	}
}

func TestAShellIsRefusedWithoutTheRightToken(t *testing.T) {
	_, err := openShell(t, client(rpcauth.Bearer("guess"))).Receive()

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func runInShell(t *testing.T, command, want string) string {
	t.Helper()
	stream := openShell(t, authed())
	err := stream.Send(&labv1.OpenShellRequest{Input: &labv1.OpenShellRequest_Keys{Keys: []byte(command)}})
	if err != nil {
		t.Fatal(err)
	}

	var output strings.Builder
	for !strings.Contains(output.String(), want) {
		res, err := stream.Receive()
		if err != nil {
			t.Fatalf("the shell ended with %v after %q", err, output.String())
		}
		output.Write(res.GetOutput())
	}
	return output.String()
}

func openShell(t *testing.T, c labv1connect.LabServiceClient) labv1connect.LabServiceOpenShellClientStream {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)
	stream, err := c.OpenShell(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stream.Close() })
	size := &labv1.TerminalSize{Cols: 120, Rows: 40}
	if err := stream.Send(&labv1.OpenShellRequest{Input: &labv1.OpenShellRequest_Size{Size: size}}); err != nil {
		t.Fatal(err)
	}
	return stream
}
