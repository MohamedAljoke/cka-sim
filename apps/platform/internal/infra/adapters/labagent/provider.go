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

type Provider struct {
	agent labv1connect.LabServiceClient
}

func NewProvider(url, token string) *Provider {
	transport := connecthttp.NewTransport(&http.Client{Timeout: 10 * time.Second}, url)
	return &Provider{agent: labv1connect.NewLabServiceClient(connect.NewClient(transport, rpcauth.Bearer(token)))}
}

func (p *Provider) Get(ctx context.Context) (labs.Lab, error) {
	res, err := p.agent.GetCluster(ctx, &labv1.GetClusterRequest{})
	if err != nil {
		return labs.Lab{}, refusal(err)
	}
	return labOf(res.GetCluster())
}

func (p *Provider) Start(ctx context.Context) (labs.Lab, error) {
	res, err := p.agent.CreateCluster(ctx, &labv1.CreateClusterRequest{})
	if err != nil {
		return labs.Lab{}, refusal(err)
	}
	return labOf(res.GetCluster())
}

func (p *Provider) Stop(ctx context.Context) (labs.Lab, error) {
	res, err := p.agent.DeleteCluster(ctx, &labv1.DeleteClusterRequest{})
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

func labOf(c *labv1.Cluster) (labs.Lab, error) {
	state, ok := states[c.GetState()]
	if !ok {
		return labs.Lab{}, fmt.Errorf("the agent answered an unknown cluster state %v", c.GetState())
	}
	lab := labs.Lab{State: state, Step: c.GetStep(), Error: c.GetError()}
	if c.GetStarted() != nil {
		lab.StartedAt = c.GetStarted().AsTime()
	}
	if c.GetReady() != nil {
		lab.ReadyAt = c.GetReady().AsTime()
	}
	return lab, nil
}
