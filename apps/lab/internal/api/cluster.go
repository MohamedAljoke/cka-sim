package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

func (a *api) GetCluster(ctx context.Context, _ *labv1.GetClusterRequest) (*labv1.GetClusterResponse, error) {
	return &labv1.GetClusterResponse{Cluster: newCluster(a.clusters.Get.Execute(ctx))}, nil
}

func (a *api) CreateCluster(ctx context.Context, _ *labv1.CreateClusterRequest) (*labv1.CreateClusterResponse, error) {
	status, err := a.clusters.Create.Execute(ctx)
	switch {
	case errors.Is(err, cluster.ErrBusy):
		return nil, connect.NewError(connect.CodeFailedPrecondition, cluster.ErrBusy.Message).WithCause(err)
	case err != nil:
		return nil, err
	}

	return &labv1.CreateClusterResponse{Cluster: newCluster(status)}, nil
}

func (a *api) DeleteCluster(ctx context.Context, _ *labv1.DeleteClusterRequest) (*labv1.DeleteClusterResponse, error) {
	status, err := a.clusters.Delete.Execute(ctx)
	switch {
	case errors.Is(err, cluster.ErrBusy):
		return nil, connect.NewError(connect.CodeFailedPrecondition, cluster.ErrBusy.Message).WithCause(err)
	case err != nil:
		return nil, err
	}

	return &labv1.DeleteClusterResponse{Cluster: newCluster(status)}, nil
}

func newCluster(s cluster.Status) *labv1.Cluster {
	msg := &labv1.Cluster{State: s.State, Step: s.Step, Error: s.Error}
	if !s.Started.IsZero() {
		msg.Started = timestamppb.New(s.Started)
	}

	if !s.Ready.IsZero() {
		msg.Ready = timestamppb.New(s.Ready)
	}

	return msg
}
