package labagent

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type agent struct {
	calls labv1connect.LabServiceClient
	work  labv1connect.LabServiceClient
}

// Shells and scripts get a client with no timeout: their own contexts say how long they may take, and a
// shell's stream needs HTTP/2.
func newAgent(url, token string) *agent {
	calls := connecthttp.NewTransport(&http.Client{Timeout: 10 * time.Second}, url)
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	streams := connecthttp.NewTransport(&http.Client{Transport: &http.Transport{Protocols: protocols}}, url)

	return &agent{
		calls: labv1connect.NewLabServiceClient(connect.NewClient(calls, rpcauth.Bearer(token))),
		work:  labv1connect.NewLabServiceClient(connect.NewClient(streams, rpcauth.Bearer(token))),
	}
}

func (p *agent) Get(ctx context.Context) (labs.Lab, error) {
	res, err := p.calls.GetCluster(ctx, &labv1.GetClusterRequest{})
	if err != nil {
		return labs.Lab{}, refusal(err)
	}

	return labOf(res.GetCluster())
}

func (p *agent) Start(ctx context.Context) (labs.Lab, error) {
	res, err := p.calls.CreateCluster(ctx, &labv1.CreateClusterRequest{})
	if err != nil {
		return labs.Lab{}, refusal(err)
	}

	return labOf(res.GetCluster())
}

func (p *agent) Stop(ctx context.Context) (labs.Lab, error) {
	res, err := p.calls.DeleteCluster(ctx, &labv1.DeleteClusterRequest{})
	if err != nil {
		return labs.Lab{}, refusal(err)
	}

	return labOf(res.GetCluster())
}

func refusal(err error) error {
	if connect.CodeOf(err) == connect.CodeFailedPrecondition {
		return fmt.Errorf("%w: %w", labs.ErrBusy, err)
	}

	return err
}

var states = map[labv1.ClusterState]labs.State{
	labv1.ClusterState_CLUSTER_STATE_ABSENT:   labs.None,
	labv1.ClusterState_CLUSTER_STATE_CREATING: labs.Starting,
	labv1.ClusterState_CLUSTER_STATE_READY:    labs.Ready,
	labv1.ClusterState_CLUSTER_STATE_DELETING: labs.Stopping,
	labv1.ClusterState_CLUSTER_STATE_FAILED:   labs.Failed,
}

func labOf(cluster *labv1.Cluster) (labs.Lab, error) {
	state, ok := states[cluster.GetState()]
	if !ok {
		return labs.Lab{}, fmt.Errorf("the agent answered an unknown cluster state %v", cluster.GetState())
	}

	lab := labs.Lab{State: state, Step: cluster.GetStep(), Error: cluster.GetError()}
	if cluster.GetStarted() != nil {
		lab.StartedAt = cluster.GetStarted().AsTime()
	}

	if cluster.GetReady() != nil {
		lab.ReadyAt = cluster.GetReady().AsTime()
	}

	return lab, nil
}
