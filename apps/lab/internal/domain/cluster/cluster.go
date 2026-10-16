package cluster

import (
	"sync"
	"time"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

const (
	Absent   = labv1.ClusterState_CLUSTER_STATE_ABSENT
	Creating = labv1.ClusterState_CLUSTER_STATE_CREATING
	Ready    = labv1.ClusterState_CLUSTER_STATE_READY
	Deleting = labv1.ClusterState_CLUSTER_STATE_DELETING
	Failed   = labv1.ClusterState_CLUSTER_STATE_FAILED
)

type Status struct {
	State   labv1.ClusterState
	Step    string
	Started time.Time
	Ready   time.Time
	Error   string
}

type Cluster struct {
	mu     sync.Mutex
	status Status
}

func New() *Cluster {
	return &Cluster{status: Status{State: Absent}}
}

func (c *Cluster) Status() Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.status
}

func (c *Cluster) BeginCreate(now time.Time) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.status.State {
	case Creating, Ready:
		return false, nil
	case Deleting:
		return false, ErrBusy
	default:
		c.status = Status{State: Creating, Started: now}

		return true, nil
	}
}

func (c *Cluster) Progress(step string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status.State == Creating {
		c.status.Step = step
	}
}

func (c *Cluster) Created(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status.State == Creating {
		c.status = Status{State: Ready, Started: c.status.Started, Ready: now}
	}
}

func (c *Cluster) CreateFailed() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status.State == Creating {
		c.status = Status{
			State:   Failed,
			Started: c.status.Started,
			Error:   "the cluster couldn't be created; the lab's log has the cause",
		}
	}
}

func (c *Cluster) BeginDelete() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.status.State {
	case Absent, Deleting:
		return false, nil
	case Creating:
		return false, ErrBusy
	default:
		c.status = Status{State: Deleting}

		return true, nil
	}
}

func (c *Cluster) Deleted() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status.State == Deleting {
		c.status = Status{State: Absent}
	}
}

func (c *Cluster) DeleteFailed() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status.State == Deleting {
		c.status = Status{State: Failed, Error: "the cluster couldn't be deleted; the lab's log has the cause"}
	}
}
