package usecases

type UseCases struct {
	List     *ListTasks
	Get      *GetTask
	Current  *GetCurrentTask
	Solution *GetSolution
	Heal     *HealCluster
	SetUp    *SetUpTask
	Start    *StartTask
	Check    *CheckTask
	Tidy     *TidyTask
	Solve    *SolveTask
	Selftest *SelftestTask
}

func New(repo TaskRepository, runner ScriptRunner) UseCases {
	clusters := newClusters()
	heal := NewHealCluster(repo, runner, clusters)
	setUp := NewSetUpTask(repo, runner, clusters)
	start := NewStartTask(repo, heal, setUp)
	check := NewCheckTask(repo, runner)
	solve := NewSolveTask(repo, runner, clusters)

	return UseCases{
		List:     NewListTasks(repo),
		Get:      NewGetTask(repo),
		Current:  NewGetCurrentTask(repo, clusters),
		Solution: NewGetSolution(repo),
		Heal:     heal,
		SetUp:    setUp,
		Start:    start,
		Check:    check,
		Tidy:     NewTidyTask(repo, runner, clusters),
		Solve:    solve,
		Selftest: NewSelftestTask(start, check, solve),
	}
}
