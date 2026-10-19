package usecases

type UseCases struct {
	Current  *GetCurrentTask
	Check    *CheckTask
	Inspect  *InspectCluster
	Solution *GetSolution
}

func New(platform Platform) UseCases {
	return UseCases{
		Current:  NewGetCurrentTask(platform),
		Check:    NewCheckTask(platform),
		Inspect:  NewInspectCluster(platform),
		Solution: NewGetSolution(platform),
	}
}
