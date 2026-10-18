package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestTidyResetsTheTaskThenDeletesItsNamespaces(t *testing.T) {
	r, c := &fakeRunner{}, newClusters()
	c.of(owner).touch("tr-kubelet")

	err := NewTidyTask(breakers, r, c).Execute(context.Background(), owner, "tr-kubelet")

	if err != nil || !slices.Equal(r.ran, []string{"fix kubelet", TidyScript("tr-kubelet")}) || r.host != "cka-sim-worker" {
		t.Errorf("got %v, ran %v on %q", err, r.ran, r.host)
	}
	if c.of(owner).isTouched("tr-kubelet") {
		t.Error("still touched after its reset")
	}
}

func TestTidyResetsATaskThatWasNotTouched(t *testing.T) {
	r := &fakeRunner{}

	NewTidyTask(breakers, r, newClusters()).Execute(context.Background(), owner, "tr-etcd")

	if !slices.Equal(r.ran, []string{"fix etcd", TidyScript("tr-etcd")}) {
		t.Errorf("ran %v", r.ran)
	}
}

func TestTidyingATaskWithoutAResetOnlyDeletesItsNamespaces(t *testing.T) {
	r := &fakeRunner{}

	err := NewTidyTask(breakers, r, newClusters()).Execute(context.Background(), owner, "wl-scale")

	if err != nil || !slices.Equal(r.ran, []string{"kubectl delete namespace -l cka-sim/task=wl-scale --wait=false >/dev/null\n"}) {
		t.Errorf("got %v, ran %v", err, r.ran)
	}
}

func TestAFailedResetKeepsTheTaskTouchedAndItsNamespaces(t *testing.T) {
	r, c := &fakeRunner{results: map[string]labs.ScriptResult{"fix etcd": {ExitCode: 1}}}, newClusters()

	err := NewTidyTask(breakers, r, c).Execute(context.Background(), owner, "tr-etcd")

	if err == nil || !c.of(owner).isTouched("tr-etcd") || len(r.ran) != 1 {
		t.Errorf("got %v, ran %v, touched %v", err, r.ran, c.of(owner).touched)
	}
}

func TestTidyingSeveralTasksGoesOnPastOneThatFails(t *testing.T) {
	r := &fakeRunner{results: map[string]labs.ScriptResult{"fix etcd": {ExitCode: 1}}}

	err := NewTidyTask(breakers, r, newClusters()).Execute(context.Background(), owner, "tr-etcd", "wl-scale")

	if err == nil || !slices.Equal(r.ran, []string{"fix etcd", TidyScript("wl-scale")}) {
		t.Errorf("got %v, ran %v", err, r.ran)
	}
}

func TestATidyOutlivesTheRequest(t *testing.T) {
	r := &fakeRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	NewTidyTask(breakers, r, newClusters()).Execute(ctx, owner, "wl-scale")

	if r.ctxErr != nil || !r.hasDeadline {
		t.Errorf("the tidy ran with ctx error %v, deadline %v", r.ctxErr, r.hasDeadline)
	}
}

func TestTidyingAnUnknownTaskIsNotFound(t *testing.T) {
	r := &fakeRunner{}

	err := NewTidyTask(breakers, r, newClusters()).Execute(context.Background(), owner, "net-svc")

	if !errors.Is(err, tasks.ErrNotFound) || len(r.ran) != 0 {
		t.Errorf("got %v, ran %v", err, r.ran)
	}
}

func TestTidyingOnALabThatIsNotReadyStaysNotReady(t *testing.T) {
	r := &fakeRunner{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}

	err := NewTidyTask(breakers, r, newClusters()).Execute(context.Background(), owner, "wl-scale")

	if !errors.Is(err, labs.ErrNotReady) {
		t.Errorf("got %v", err)
	}
}

func TestTidyingOnALabThatDoesNotAnswerIsUnavailable(t *testing.T) {
	r := &fakeRunner{err: errors.New("connection refused")}

	err := NewTidyTask(breakers, r, newClusters()).Execute(context.Background(), owner, "wl-scale")

	if !errors.Is(err, labs.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}
