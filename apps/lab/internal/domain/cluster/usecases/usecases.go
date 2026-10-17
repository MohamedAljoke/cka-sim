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
	RunScript *RunScript
}

func New(lab *cluster.Cluster, p Provisioner, o ShellOpener, r ScriptRunner, now func() time.Time) UseCases {
	return UseCases{
		Get:       NewGetCluster(lab),
		Create:    NewCreateCluster(lab, p, now),
		Delete:    NewDeleteCluster(lab, p),
		OpenShell: NewOpenShell(lab, o),
		RunScript: NewRunScript(lab, p, r),
	}
}
