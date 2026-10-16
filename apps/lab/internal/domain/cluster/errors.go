package cluster

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var ErrBusy = domainerr.New("cluster_busy", "the cluster is still being created or deleted; wait for it to finish")
