package labagent

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"google.golang.org/protobuf/types/known/timestamppb"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

const token = "s3cret"

var started = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)

func TestEveryClusterStateBecomesALabState(t *testing.T) {
	for state, want := range map[labv1.ClusterState]labs.State{
		labv1.ClusterState_CLUSTER_STATE_ABSENT:   labs.None,
		labv1.ClusterState_CLUSTER_STATE_CREATING: labs.Starting,
		labv1.ClusterState_CLUSTER_STATE_READY:    labs.Ready,
		labv1.ClusterState_CLUSTER_STATE_DELETING: labs.Stopping,
		labv1.ClusterState_CLUSTER_STATE_FAILED:   labs.Failed,
	} {
		p := NewProvider(serve(t, &fakeAgent{cluster: &labv1.Cluster{State: state}}), token)

		got, err := p.Get(context.Background())

		if err != nil || got.State != want {
			t.Errorf("%v: got %+v, %v; want %s", state, got, err, want)
		}
	}
}

func TestALabCarriesTheClustersStepAndTimes(t *testing.T) {
	agent := &fakeAgent{cluster: &labv1.Cluster{
		State:   labv1.ClusterState_CLUSTER_STATE_READY,
		Started: timestamppb.New(started),
		Ready:   timestamppb.New(started.Add(time.Minute)),
	}}

	got, err := NewProvider(serve(t, agent), token).Start(context.Background())

	if err != nil || !got.StartedAt.Equal(started) || !got.ReadyAt.Equal(started.Add(time.Minute)) {
		t.Errorf("got %+v, %v", got, err)
	}
	if agent.creates != 1 {
		t.Errorf("CreateCluster ran %d times", agent.creates)
	}
}

func TestStopDeletesTheCluster(t *testing.T) {
	agent := &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_DELETING, Step: ""}}

	got, err := NewProvider(serve(t, agent), token).Stop(context.Background())

	if err != nil || got.State != labs.Stopping || agent.deletes != 1 {
		t.Errorf("got %+v, %v; DeleteCluster ran %d times", got, err, agent.deletes)
	}
}

func TestAFailedPreconditionIsABusyLab(t *testing.T) {
	agent := &fakeAgent{err: connect.NewError(connect.CodeFailedPrecondition, "the cluster is still being created")}

	_, err := NewProvider(serve(t, agent), token).Stop(context.Background())

	if !errors.Is(err, labs.ErrBusy) {
		t.Errorf("got %v, want ErrBusy", err)
	}
}

func TestAnUnknownStateIsAnError(t *testing.T) {
	agent := &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_UNSPECIFIED}}

	if _, err := NewProvider(serve(t, agent), token).Get(context.Background()); err == nil {
		t.Error("an unspecified state was taken as a lab")
	}
}

func TestAWrongTokenOrNoAgentIsAnErrorButNotABusyLab(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := "http://" + ln.Addr().String()
	ln.Close()

	for name, p := range map[string]*Provider{
		"wrong token": NewProvider(serve(t, &fakeAgent{cluster: &labv1.Cluster{}}), "guess"),
		"no agent":    NewProvider(closed, token),
	} {
		_, err := p.Get(context.Background())

		if err == nil || errors.Is(err, labs.ErrBusy) {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

type fakeAgent struct {
	labv1connect.UnimplementedLabServiceHandler
	cluster *labv1.Cluster
	err     error
	creates int
	deletes int
}

func (a *fakeAgent) GetCluster(context.Context, *labv1.GetClusterRequest) (*labv1.GetClusterResponse, error) {
	return &labv1.GetClusterResponse{Cluster: a.cluster}, a.err
}

func (a *fakeAgent) CreateCluster(context.Context, *labv1.CreateClusterRequest) (*labv1.CreateClusterResponse, error) {
	a.creates++
	return &labv1.CreateClusterResponse{Cluster: a.cluster}, a.err
}

func (a *fakeAgent) DeleteCluster(context.Context, *labv1.DeleteClusterRequest) (*labv1.DeleteClusterResponse, error) {
	a.deletes++
	return &labv1.DeleteClusterResponse{Cluster: a.cluster}, a.err
}

func serve(t *testing.T, agent *fakeAgent) string {
	t.Helper()
	server := connect.NewServer(rpcauth.RequireBearer(token))
	labv1connect.RegisterLabServiceHandler(server, agent)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}
