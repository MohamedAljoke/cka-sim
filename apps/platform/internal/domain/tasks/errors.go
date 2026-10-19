package tasks

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrNotFound       = domainerr.New("task_not_found", "no task has that id")
	ErrUnavailable    = domainerr.New("tasks_unavailable", "the tasks can't be loaded right now; try again")
	ErrNoCurrentTask  = domainerr.New("no_current_task", "no task is set up; open one and press Set up task")
	ErrMalformedTopic = domainerr.New("malformed_topic", "a topic is lowercase-kebab, like rolling-update")
	ErrUnknownDomain  = domainerr.New("unknown_domain",
		"the domain must be troubleshooting, architecture, networking, workloads or storage")
)
