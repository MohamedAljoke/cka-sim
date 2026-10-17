package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect/v2"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

const (
	controlPlane = clusterName + "-control-plane"
	worker       = clusterName + "-worker"
)

func TestAScriptNeedsAReadyCluster(t *testing.T) {
	if *create {
		t.Skip("the -create run changes the cluster")
	}

	_, err := authed().RunScript(context.Background(), &labv1.RunScriptRequest{Host: controlPlane, Script: "true"})

	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("got %v", err)
	}
}

// runScripts needs the ready cluster TestARealClusterIsBuiltAndDeleted builds.
func runScripts(t *testing.T) {
	t.Helper()
	c := authed()

	got, err := c.RunScript(context.Background(), &labv1.RunScriptRequest{
		Host:   controlPlane,
		Script: "echo $((40 + 2))\nkubectl get namespace default -o name\necho oops >&2\nexit 3\n",
	})
	if err != nil || got.ExitCode != 3 {
		t.Fatalf("RunScript = %v, %v", got, err)
	}
	for _, want := range []string{"42\n", "namespace/default\n", "oops\n"} {
		if !strings.Contains(string(got.Output), want) {
			t.Errorf("output %q has no %q", got.Output, want)
		}
	}

	got, err = c.RunScript(context.Background(), &labv1.RunScriptRequest{Host: worker, Script: "kubectl get namespace default -o name\n"})
	if err != nil || got.ExitCode != 0 {
		t.Errorf("kubectl on a worker = %q, %v", got.GetOutput(), err)
	}

	_, err = c.RunScript(context.Background(), &labv1.RunScriptRequest{Host: clusterName + "-base", Script: "true"})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("a host outside the cluster: got %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := c.RunScript(ctx, &labv1.RunScriptRequest{Host: controlPlane, Script: "sleep 4242\n"}); err == nil {
		t.Fatal("a script past its deadline answered")
	}
	time.Sleep(2 * time.Second)
	left, err := c.RunScript(context.Background(), &labv1.RunScriptRequest{Host: controlPlane, Script: "pgrep -f 'sleep 4242'\n"})
	if err != nil || left.ExitCode == 0 {
		t.Errorf("a stopped script is still running: %q, %v", left.GetOutput(), err)
	}
}
