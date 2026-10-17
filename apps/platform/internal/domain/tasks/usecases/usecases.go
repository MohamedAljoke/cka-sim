package usecases

type UseCases struct {
	List     *ListTasks
	Get      *GetTask
	Solution *GetSolution
	Start    *StartTask
	Check    *CheckTask
}

func New(repo TaskRepository, runner ScriptRunner) UseCases {
	return UseCases{
		List:     NewListTasks(repo),
		Get:      NewGetTask(repo),
		Solution: NewGetSolution(repo),
		Start:    NewStartTask(repo, runner),
		Check:    NewCheckTask(repo, runner),
	}
}
