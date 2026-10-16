package labs

import "time"

type State string

const (
	None     State = "none"
	Starting State = "starting"
	Ready    State = "ready"
	Stopping State = "stopping"
	Failed   State = "failed"
)

type Lab struct {
	State     State
	Step      string
	StartedAt time.Time
	ReadyAt   time.Time
	Error     string
}
