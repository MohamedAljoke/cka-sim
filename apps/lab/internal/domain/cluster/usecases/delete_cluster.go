package usecases

import (
	"context"
	"log"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

const deleteTimeout = 5 * time.Minute

type DeleteCluster struct {
	cluster     *cluster.Cluster
	provisioner Provisioner
}

func NewDeleteCluster(c *cluster.Cluster, p Provisioner) *DeleteCluster {
	return &DeleteCluster{cluster: c, provisioner: p}
}

func (uc *DeleteCluster) Execute(ctx context.Context) (cluster.Status, error) {
	start, err := uc.cluster.BeginDelete()
	if err != nil {
		return cluster.Status{}, err
	}

	if start {
		go uc.delete(context.WithoutCancel(ctx))
	}

	return uc.cluster.Status(), nil
}

func (uc *DeleteCluster) delete(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	if err := uc.provisioner.Delete(ctx); err != nil {
		log.Printf("delete the cluster: %v", err)
		uc.cluster.DeleteFailed()

		return
	}

	uc.cluster.Deleted()
}
