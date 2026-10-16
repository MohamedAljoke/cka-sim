package usecases

type UseCases struct {
	List     *ListTasks
	Get      *GetTask
	Solution *GetSolution
}

func New(repo TaskRepository) UseCases {
	return UseCases{List: NewListTasks(repo), Get: NewGetTask(repo), Solution: NewGetSolution(repo)}
}
