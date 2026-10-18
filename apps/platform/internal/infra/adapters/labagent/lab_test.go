package labagent

import (
	"context"
	"errors"
	"testing"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

func TestAFixedMachinesLabIsWhatItsAgentSays(t *testing.T) {
	agent := &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_ABSENT}}
	lab := NewLab(Fixed(serve(t, agent)), token)

	got, err := lab.Get(context.Background())
	if err != nil || got.State != labs.None {
		t.Errorf("got %+v, %v; want none", got, err)
	}
	if agent.creates != 0 {
		t.Errorf("a fixed machine's agent was told to build a cluster %d times", agent.creates)
	}
}

func TestAnOnDemandAgentWithNoClusterIsToldToBuildOne(t *testing.T) {
	agent := &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_ABSENT}}
	machine := &fakeMachine{place: Place{URL: serve(t, agent), OnDemand: true}}

	if _, err := NewLab(machine, token).Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	if agent.creates != 1 {
		t.Errorf("CreateCluster ran %d times, want 1", agent.creates)
	}
}

func TestAMachineWithNoAgentYetAnswersItsOwnLab(t *testing.T) {
	waiting := labs.Lab{State: labs.Starting, Step: "Waiting for a node"}
	machine := &fakeMachine{place: Place{Lab: waiting}}

	got, err := NewLab(machine, token).Get(context.Background())
	if err != nil || got != waiting {
		t.Errorf("got %+v, %v; want %+v", got, err, waiting)
	}
}

func TestStartBringsTheMachineUpAndStartsItsAgent(t *testing.T) {
	agent := &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING}}
	machine := &fakeMachine{place: Place{URL: serve(t, agent), OnDemand: true}}

	got, err := NewLab(machine, token).Start(context.Background())
	if err != nil || got.State != labs.Starting {
		t.Errorf("got %+v, %v; want starting", got, err)
	}
	if machine.ups != 1 || agent.creates != 1 {
		t.Errorf("Up ran %d times and CreateCluster %d, want 1 each", machine.ups, agent.creates)
	}
}

func TestStartWhileTheMachineComesUpAnswersItsLab(t *testing.T) {
	waiting := labs.Lab{State: labs.Starting, Step: "Pulling the lab image"}
	machine := &fakeMachine{place: Place{Lab: waiting}}

	got, err := NewLab(machine, token).Start(context.Background())
	if err != nil || got != waiting || machine.ups != 1 {
		t.Errorf("got %+v, %v after %d Ups; want %+v after 1", got, err, machine.ups, waiting)
	}
}

func TestStartWhileTheLastLabStopsIsBusy(t *testing.T) {
	machine := &fakeMachine{place: Place{Lab: labs.Lab{State: labs.Stopping}}}

	if _, err := NewLab(machine, token).Start(context.Background()); !errors.Is(err, labs.ErrBusy) {
		t.Errorf("got %v, want ErrBusy", err)
	}
}

func TestStartAnswersAMachineThatCantComeUp(t *testing.T) {
	broken := errors.New("pods is forbidden")
	machine := &fakeMachine{upErr: broken}

	if _, err := NewLab(machine, token).Start(context.Background()); !errors.Is(err, broken) {
		t.Errorf("got %v, want %v", err, broken)
	}
}

func TestStopRemovesAnOnDemandMachineWithoutAskingItsAgent(t *testing.T) {
	agent := &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY}}
	machine := &fakeMachine{place: Place{URL: serve(t, agent), OnDemand: true}, removes: true}

	got, err := NewLab(machine, token).Stop(context.Background())
	if err != nil || got.State != labs.Stopping {
		t.Errorf("got %+v, %v; want stopping", got, err)
	}
	if machine.downs != 1 || agent.deletes != 0 {
		t.Errorf("Down ran %d times and DeleteCluster %d, want 1 and 0", machine.downs, agent.deletes)
	}
}

func TestStopOnAFixedMachineDeletesTheCluster(t *testing.T) {
	agent := &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_DELETING}}

	got, err := NewLab(Fixed(serve(t, agent)), token).Stop(context.Background())
	if err != nil || got.State != labs.Stopping || agent.deletes != 1 {
		t.Errorf("got %+v, %v; DeleteCluster ran %d times", got, err, agent.deletes)
	}
}

func TestScriptsAndShellsWaitForAnAgent(t *testing.T) {
	lab := NewLab(&fakeMachine{place: Place{Lab: labs.Lab{State: labs.Starting}}}, token)

	if _, err := lab.RunScript(context.Background(), "cka-sim-worker", "true"); !errors.Is(err, labs.ErrNotReady) {
		t.Errorf("RunScript: got %v, want ErrNotReady", err)
	}
	if _, err := lab.OpenShell(context.Background(), labs.TerminalSize{Cols: 80, Rows: 24}); !errors.Is(err, labs.ErrNotReady) {
		t.Errorf("OpenShell: got %v, want ErrNotReady", err)
	}
}

func TestANewMachineIsReachedAtItsNewAddress(t *testing.T) {
	first := &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY}}
	second := &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING}}
	machine := &fakeMachine{place: Place{URL: serve(t, first)}}
	lab := NewLab(machine, token)

	if got, err := lab.Get(context.Background()); err != nil || got.State != labs.Ready {
		t.Fatalf("first machine: got %+v, %v", got, err)
	}
	machine.place.URL = serve(t, second)
	if got, err := lab.Get(context.Background()); err != nil || got.State != labs.Starting {
		t.Errorf("second machine: got %+v, %v; want starting", got, err)
	}
}

type fakeMachine struct {
	place   Place
	upErr   error
	removes bool
	ups     int
	downs   int
}

func (m *fakeMachine) Up(context.Context) error {
	m.ups++
	return m.upErr
}

func (m *fakeMachine) Down(context.Context) (bool, error) {
	m.downs++
	return m.removes, nil
}

func (m *fakeMachine) Where(context.Context) (Place, error) {
	return m.place, nil
}
