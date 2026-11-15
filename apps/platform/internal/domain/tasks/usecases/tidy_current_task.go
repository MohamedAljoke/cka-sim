package usecases

import "context"

type TidyCurrentTask struct {
	tidy     *TidyTask
	clusters *clusters
}

func NewTidyCurrentTask(tidy *TidyTask, clusters *clusters) *TidyCurrentTask {
	return &TidyCurrentTask{tidy: tidy, clusters: clusters}
}

// Execute tidies and leaves the task the owner is on, if any, for a caller that takes the cluster over.
func (uc *TidyCurrentTask) Execute(ctx context.Context, owner string) error {
	id := uc.clusters.of(owner).currentTask()
	if id == "" {
		return nil
	}

	return uc.tidy.Execute(ctx, owner, id)
}
