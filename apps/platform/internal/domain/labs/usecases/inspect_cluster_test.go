package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

func TestInspectRunsTheCommandOnTheOwnersControlPlane(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "NAME READY\nweb  1/1\n"}}

	got, err := NewInspectCluster(r).Execute(context.Background(), "alice", []string{"get", "pods"})

	if err != nil || got.Output != "NAME READY\nweb  1/1\n" || got.ExitCode != 0 || got.Truncated {
		t.Fatalf("Execute = %+v, %v", got, err)
	}
	if r.owner != "alice" || r.host != labs.ControlPlane || r.script != "timeout 15 kubectl 'get' 'pods'\n" || !r.hasDeadline {
		t.Errorf("ran %q on %q for %q, deadline %v", r.script, r.host, r.owner, r.hasDeadline)
	}
}

func TestAFailedCommandIsAnAnswerNotAnError(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: `Error from server (NotFound): pods "web" not found`, ExitCode: 1}}

	got, err := NewInspectCluster(r).Execute(context.Background(), "alice", []string{"get", "pod", "web"})

	if err != nil || got.ExitCode != 1 || !strings.Contains(got.Output, "NotFound") {
		t.Errorf("Execute = %+v, %v", got, err)
	}
}

func TestALongOutputKeepsItsStart(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "first line\n" + strings.Repeat("x", inspectOutputLimit)}}

	got, err := NewInspectCluster(r).Execute(context.Background(), "alice", []string{"get", "pods"})

	if err != nil || !got.Truncated || len(got.Output) != inspectOutputLimit || !strings.HasPrefix(got.Output, "first line\n") {
		t.Errorf("Execute = %d bytes, truncated %v, %v", len(got.Output), got.Truncated, err)
	}
}

func TestACommandThatIsNotReadOnlyNeverRuns(t *testing.T) {
	r := &fakeRunner{}

	_, err := NewInspectCluster(r).Execute(context.Background(), "alice", []string{"delete", "pod", "web"})

	if !errors.Is(err, labs.ErrNotReadOnly) || r.script != "" {
		t.Errorf("got %v, ran %q", err, r.script)
	}
}

func TestInspectOnALabThatIsNotReadyStaysNotReady(t *testing.T) {
	r := &fakeRunner{err: fmt.Errorf("%w: creating", labs.ErrNotReady)}

	_, err := NewInspectCluster(r).Execute(context.Background(), "alice", []string{"get", "pods"})

	if !errors.Is(err, labs.ErrNotReady) || errors.Is(err, labs.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestInspectOnALabThatDoesNotAnswerIsUnavailableAndKeepsTheCause(t *testing.T) {
	cause := errors.New("dial tcp 127.0.0.1:7171: connection refused")

	_, err := NewInspectCluster(&fakeRunner{err: cause}).Execute(context.Background(), "alice", []string{"get", "pods"})

	if !errors.Is(err, labs.ErrUnavailable) || !errors.Is(err, cause) {
		t.Errorf("got %v", err)
	}
}

type fakeRunner struct {
	result      labs.ScriptResult
	err         error
	owner       string
	host        string
	script      string
	hasDeadline bool
}

func (r *fakeRunner) RunScript(ctx context.Context, owner, host, script string) (labs.ScriptResult, error) {
	r.owner, r.host, r.script = owner, host, script
	_, r.hasDeadline = ctx.Deadline()
	return r.result, r.err
}
