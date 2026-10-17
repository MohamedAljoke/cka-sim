package usecases

import (
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

type UseCases struct {
	Get       *GetCluster
	Create    *CreateCluster
	Delete    *DeleteCluster
	OpenShell *OpenShell
}

func New(c *cluster.Cluster, p Provisioner, o ShellOpener, now func() time.Time) UseCases {
	return UseCases{
		Get:       NewGetCluster(c),
		Create:    NewCreateCluster(c, p, now),
		Delete:    NewDeleteCluster(c, p),
		OpenShell: NewOpenShell(c, o),
	}
}
