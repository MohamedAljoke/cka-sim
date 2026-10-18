package usecases

type UseCases struct {
	List     *ListTasks
	Get      *GetTask
	Solution *GetSolution
	Heal     *HealCluster
	SetUp    *SetUpTask
	Start    *StartTask
	Check    *CheckTask
	Tidy     *TidyTask
	Selftest *SelftestTask
}

func New(repo TaskRepository, runner ScriptRunner) UseCases {
	cluster := newCluster()
	heal := NewHealCluster(repo, runner, cluster)
	setUp := NewSetUpTask(repo, runner, cluster)
	start := NewStartTask(repo, heal, setUp)
	check := NewCheckTask(repo, runner)

	return UseCases{
		List:     NewListTasks(repo),
		Get:      NewGetTask(repo),
		Solution: NewGetSolution(repo),
		Heal:     heal,
		SetUp:    setUp,
		Start:    start,
		Check:    check,
		Tidy:     NewTidyTask(repo, runner, cluster),
		Selftest: NewSelftestTask(repo, runner, cluster, start, check),
	}
}
