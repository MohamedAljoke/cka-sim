package tasks

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrNotFound    = domainerr.New("task_not_found", "no task has that id")
	ErrUnavailable = domainerr.New("tasks_unavailable", "the tasks can't be loaded right now; try again")
)
