package usecases

import (
	"context"
	"log"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

const createTimeout = 20 * time.Minute

type CreateCluster struct {
	cluster     *cluster.Cluster
	provisioner Provisioner
	now         func() time.Time
}

func NewCreateCluster(c *cluster.Cluster, p Provisioner, now func() time.Time) *CreateCluster {
	return &CreateCluster{cluster: c, provisioner: p, now: now}
}

func (uc *CreateCluster) Execute(ctx context.Context) (cluster.Status, error) {
	start, err := uc.cluster.BeginCreate(uc.now())
	if err != nil {
		return cluster.Status{}, err
	}

	if start {
		go uc.create(context.WithoutCancel(ctx))
	}

	return uc.cluster.Status(), nil
}

func (uc *CreateCluster) create(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	if err := uc.provisioner.Create(ctx, uc.cluster.Progress); err != nil {
		log.Printf("create the cluster: %v", err)
		uc.cluster.CreateFailed()

		return
	}

	uc.cluster.Created(uc.now())
}
