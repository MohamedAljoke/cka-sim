package usecases

import "context"

type GetSolution struct {
	platform Platform
}

func NewGetSolution(platform Platform) *GetSolution {
	return &GetSolution{platform: platform}
}

// Execute answers taskID's solution, or the solution of the task the student is on when taskID is empty.
func (uc *GetSolution) Execute(ctx context.Context, student, taskID string) (string, error) {
	taskID, err := orCurrent(ctx, uc.platform, student, taskID)
	if err != nil {
		return "", err
	}

	return uc.platform.Solution(ctx, student, taskID)
}
