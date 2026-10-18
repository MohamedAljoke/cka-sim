package labagent

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect/v2"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

func TestAScriptRunsOnTheAgentAndAnswersItsResult(t *testing.T) {
	agent := &fakeAgent{script: &labv1.RunScriptResponse{Output: []byte("no such CRD\n"), ExitCode: 1}}

	got, err := newAgent(serve(t, agent), token).RunScript(context.Background(), "cka-sim-control-plane", "kubectl get widgets\n")

	if err != nil || got.Output != "no such CRD\n" || got.ExitCode != 1 || got.Passed() {
		t.Fatalf("got %+v, %v", got, err)
	}
	if agent.ran.GetHost() != "cka-sim-control-plane" || agent.ran.GetScript() != "kubectl get widgets\n" {
		t.Errorf("the agent ran %v", agent.ran)
	}
}

func TestAScriptOnAClusterThatIsNotReadyIsNotReady(t *testing.T) {
	agent := &fakeAgent{err: connect.NewError(connect.CodeFailedPrecondition, "the cluster isn't ready")}

	_, err := newAgent(serve(t, agent), token).RunScript(context.Background(), "cka-sim-control-plane", "true\n")

	if !errors.Is(err, labs.ErrNotReady) {
		t.Errorf("got %v, want ErrNotReady", err)
	}
}

func TestAScriptOnAnUnknownHostIsAnErrorButNotNotReady(t *testing.T) {
	agent := &fakeAgent{err: connect.NewError(connect.CodeInvalidArgument, "the host isn't one of the cluster's nodes")}

	_, err := newAgent(serve(t, agent), token).RunScript(context.Background(), "my-database", "true\n")

	if err == nil || errors.Is(err, labs.ErrNotReady) {
		t.Errorf("got %v", err)
	}
}
