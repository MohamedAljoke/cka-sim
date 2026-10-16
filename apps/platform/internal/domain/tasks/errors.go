package tasks

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrNotFound    = domainerr.NotFound("task_not_found", "no task has that id")
	ErrUnavailable = domainerr.Unavailable("tasks_unavailable", "the tasks can't be loaded right now; try again").AsFault()
)
