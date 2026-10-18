package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

var (
	breaksKubelet = tasks.Task{ID: "tr-kubelet", Host: "cka-sim-worker", Setup: "break kubelet", Reset: "fix kubelet"}
	breaksEtcd    = tasks.Task{ID: "tr-etcd", Host: "cka-sim-control-plane", Setup: "break etcd", Reset: "fix etcd"}
	breakers      = fakeRepository{tasks: []tasks.Task{breaksKubelet, breaksEtcd, withSetup}}
)

func TestHealResetsOnlyTheTasksSetUpSinceTheirLastReset(t *testing.T) {
	r, c := &fakeRunner{}, newCluster()
	c.touch("tr-etcd")
	c.touch("wl-scale")

	got, err := NewHealCluster(breakers, r, c).Execute(context.Background())

	if err != nil || !got.Passed() || !slices.Equal(r.ran, []string{"fix etcd"}) || r.host != "cka-sim-control-plane" {
		t.Errorf("Execute = %+v, %v; ran %v on %q", got, err, r.ran, r.host)
	}
	if c.isTouched("tr-etcd") || c.isTouched("wl-scale") {
		t.Errorf("still touched: %v", c.touched)
	}
}

func TestAHealedClusterHealsNothingTheNextTime(t *testing.T) {
	r, c := &fakeRunner{}, newCluster()
	c.touch("tr-kubelet")
	heal := NewHealCluster(breakers, r, c)

	heal.Execute(context.Background())
	heal.Execute(context.Background())

	if !slices.Equal(r.ran, []string{"fix kubelet"}) {
		t.Errorf("ran %v", r.ran)
	}
}

func TestAFailedResetAnswersAndIsTriedAgain(t *testing.T) {
	r, c := &fakeRunner{results: map[string]labs.ScriptResult{"fix kubelet": {ExitCode: 1, Output: "kubelet: no such file\n"}}}, newCluster()
	c.touch("tr-kubelet")
	c.touch("tr-etcd")
	heal := NewHealCluster(breakers, r, c)

	got, err := heal.Execute(context.Background())

	if err != nil || got.ExitCode != 1 || !strings.Contains(got.Output, "tr-kubelet") || !strings.Contains(got.Output, "no such file") {
		t.Errorf("Execute = %+v, %v", got, err)
	}
	if !c.isTouched("tr-kubelet") || c.isTouched("tr-etcd") {
		t.Errorf("touched %v", c.touched)
	}

	delete(r.results, "fix kubelet")
	heal.Execute(context.Background())

	if !slices.Equal(r.ran, []string{"fix kubelet", "fix etcd", "fix kubelet"}) || c.isTouched("tr-kubelet") {
		t.Errorf("ran %v, touched %v", r.ran, c.touched)
	}
}

func TestAResetOutlivesTheRequestButNotItsTimeout(t *testing.T) {
	r, c := &fakeRunner{}, newCluster()
	c.touch("tr-etcd")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	NewHealCluster(breakers, r, c).Execute(ctx)

	if r.ctxErr != nil || !r.hasDeadline {
		t.Errorf("the reset ran with ctx error %v, deadline %v", r.ctxErr, r.hasDeadline)
	}
}

func TestHealingALabThatIsNotReadyStaysNotReady(t *testing.T) {
	r, c := &fakeRunner{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}, newCluster()
	c.touch("tr-etcd")

	_, err := NewHealCluster(breakers, r, c).Execute(context.Background())

	if !errors.Is(err, labs.ErrNotReady) || !c.isTouched("tr-etcd") {
		t.Errorf("got %v, touched %v", err, c.touched)
	}
}

func TestHealingALabThatDoesNotAnswerIsUnavailable(t *testing.T) {
	r, c := &fakeRunner{err: errors.New("connection refused")}, newCluster()
	c.touch("tr-etcd")

	_, err := NewHealCluster(breakers, r, c).Execute(context.Background())

	if !errors.Is(err, labs.ErrUnavailable) || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("got %v", err)
	}
}

func TestHealIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	_, err := NewHealCluster(fakeRepository{err: errors.New("disk gone")}, &fakeRunner{}, newCluster()).Execute(context.Background())

	if !errors.Is(err, tasks.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}
