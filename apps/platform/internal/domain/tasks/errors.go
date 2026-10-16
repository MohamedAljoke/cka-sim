package tasks

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var ErrUnavailable = domainerr.Unavailable("tasks_unavailable", "the tasks can't be loaded right now; try again").AsFault()
