package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

type GetCluster struct {
	cluster *cluster.Cluster
}

func NewGetCluster(c *cluster.Cluster) *GetCluster {
	return &GetCluster{cluster: c}
}

func (uc *GetCluster) Execute(context.Context) cluster.Status {
	return uc.cluster.Status()
}
