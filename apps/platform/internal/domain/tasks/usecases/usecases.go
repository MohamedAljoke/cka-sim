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
}

func New(repo TaskRepository, runner ScriptRunner) UseCases {
	cluster := newCluster()
	heal := NewHealCluster(repo, runner, cluster)
	setUp := NewSetUpTask(repo, runner, cluster)

	return UseCases{
		List:     NewListTasks(repo),
		Get:      NewGetTask(repo),
		Solution: NewGetSolution(repo),
		Heal:     heal,
		SetUp:    setUp,
		Start:    NewStartTask(repo, heal, setUp),
		Check:    NewCheckTask(repo, runner),
		Tidy:     NewTidyTask(repo, runner, cluster),
	}
}
