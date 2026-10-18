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
	r, c := &fakeRunner{}, newClusters()
	c.of(owner).touch("tr-etcd")
	c.of(owner).touch("wl-scale")

	got, err := NewHealCluster(breakers, r, c).Execute(context.Background(), owner)

	if err != nil || !got.Passed() || !slices.Equal(r.ran, []string{"fix etcd"}) || r.host != "cka-sim-control-plane" {
		t.Errorf("Execute = %+v, %v; ran %v on %q", got, err, r.ran, r.host)
	}
	if c.of(owner).isTouched("tr-etcd") || c.of(owner).isTouched("wl-scale") {
		t.Errorf("still touched: %v", c.of(owner).touched)
	}
}

func TestAHealedClusterHealsNothingTheNextTime(t *testing.T) {
	r, c := &fakeRunner{}, newClusters()
	c.of(owner).touch("tr-kubelet")
	heal := NewHealCluster(breakers, r, c)

	heal.Execute(context.Background(), owner)
	heal.Execute(context.Background(), owner)

	if !slices.Equal(r.ran, []string{"fix kubelet"}) {
		t.Errorf("ran %v", r.ran)
	}
}

func TestAFailedResetAnswersAndIsTriedAgain(t *testing.T) {
	r, c := &fakeRunner{results: map[string]labs.ScriptResult{"fix kubelet": {ExitCode: 1, Output: "kubelet: no such file\n"}}}, newClusters()
	c.of(owner).touch("tr-kubelet")
	c.of(owner).touch("tr-etcd")
	heal := NewHealCluster(breakers, r, c)

	got, err := heal.Execute(context.Background(), owner)

	if err != nil || got.ExitCode != 1 || !strings.Contains(got.Output, "tr-kubelet") || !strings.Contains(got.Output, "no such file") {
		t.Errorf("Execute = %+v, %v", got, err)
	}
	if !c.of(owner).isTouched("tr-kubelet") || c.of(owner).isTouched("tr-etcd") {
		t.Errorf("touched %v", c.of(owner).touched)
	}

	delete(r.results, "fix kubelet")
	heal.Execute(context.Background(), owner)

	if !slices.Equal(r.ran, []string{"fix kubelet", "fix etcd", "fix kubelet"}) || c.of(owner).isTouched("tr-kubelet") {
		t.Errorf("ran %v, touched %v", r.ran, c.of(owner).touched)
	}
}

func TestAResetOutlivesTheRequestButNotItsTimeout(t *testing.T) {
	r, c := &fakeRunner{}, newClusters()
	c.of(owner).touch("tr-etcd")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	NewHealCluster(breakers, r, c).Execute(ctx, owner)

	if r.ctxErr != nil || !r.hasDeadline {
		t.Errorf("the reset ran with ctx error %v, deadline %v", r.ctxErr, r.hasDeadline)
	}
}

func TestHealingALabThatIsNotReadyStaysNotReady(t *testing.T) {
	r, c := &fakeRunner{err: fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady)}, newClusters()
	c.of(owner).touch("tr-etcd")

	_, err := NewHealCluster(breakers, r, c).Execute(context.Background(), owner)

	if !errors.Is(err, labs.ErrNotReady) || !c.of(owner).isTouched("tr-etcd") {
		t.Errorf("got %v, touched %v", err, c.of(owner).touched)
	}
}

func TestHealingALabThatDoesNotAnswerIsUnavailable(t *testing.T) {
	r, c := &fakeRunner{err: errors.New("connection refused")}, newClusters()
	c.of(owner).touch("tr-etcd")

	_, err := NewHealCluster(breakers, r, c).Execute(context.Background(), owner)

	if !errors.Is(err, labs.ErrUnavailable) || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("got %v", err)
	}
}

func TestHealIsUnavailableWhenTheRepositoryFails(t *testing.T) {
	_, err := NewHealCluster(fakeRepository{err: errors.New("disk gone")}, &fakeRunner{}, newClusters()).Execute(context.Background(), owner)

	if !errors.Is(err, tasks.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}
