package cluster

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrBusy     = domainerr.New("cluster_busy", "the cluster is still being created or deleted; wait for it to finish")
	ErrNotReady = domainerr.New("cluster_not_ready", "the cluster isn't ready; create it and wait for it first")
)
