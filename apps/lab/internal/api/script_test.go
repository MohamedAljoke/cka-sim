package api

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect/v2"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

func TestRunScriptAnswersTheOutputAndExitCode(t *testing.T) {
	r := &fakeRunner{result: cluster.ScriptResult{Output: []byte("namespace/ar-aggregate created\n"), ExitCode: 3}}
	c := readyLab(t, newEchoOpener(), r)

	got, err := c.RunScript(context.Background(), &labv1.RunScriptRequest{Host: "lab-worker", Script: "kubectl create ns ar-aggregate\nexit 3\n"})

	if err != nil || string(got.Output) != "namespace/ar-aggregate created\n" || got.ExitCode != 3 {
		t.Fatalf("got %v, %v", got, err)
	}
	if r.host != "lab-worker" || r.script != "kubectl create ns ar-aggregate\nexit 3\n" {
		t.Errorf("ran %q on %q", r.script, r.host)
	}
}

func TestRunScriptNeedsAReadyCluster(t *testing.T) {
	c := client(serve(t, &fakeProvisioner{}))

	_, err := c.RunScript(context.Background(), &labv1.RunScriptRequest{Host: "lab-control-plane", Script: "true"})

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeFailedPrecondition || connectErr.Message() != cluster.ErrNotReady.Message {
		t.Errorf("got %v", err)
	}
}

func TestRunScriptRefusesAHostOutsideTheCluster(t *testing.T) {
	r := &fakeRunner{}
	c := readyLab(t, newEchoOpener(), r)

	_, err := c.RunScript(context.Background(), &labv1.RunScriptRequest{Host: "my-database", Script: "rm -rf /"})

	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("got %v", err)
	}
	if r.host != "" {
		t.Errorf("ran a script on %q", r.host)
	}
}

func TestRunScriptNeedsTheToken(t *testing.T) {
	c := shellClient(serve(t, &fakeProvisioner{}), rpcauth.Bearer("guess"))

	_, err := c.RunScript(context.Background(), &labv1.RunScriptRequest{Host: "lab-control-plane", Script: "true"})

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

type fakeRunner struct {
	result cluster.ScriptResult
	host   string
	script string
}

func (r *fakeRunner) Run(_ context.Context, host, script string) (cluster.ScriptResult, error) {
	r.host, r.script = host, script
	return r.result, nil
}
