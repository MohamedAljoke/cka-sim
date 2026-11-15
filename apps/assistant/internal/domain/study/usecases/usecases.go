package usecases

type UseCases struct {
	List    *ListTasks
	Topics  *ListTopics
	Start   *StartTask
	History *GetHistory
	Note    *SaveNote
}

func New(platform Platform) UseCases {
	list := NewListTasks(platform)

	return UseCases{
		List:    list,
		Topics:  NewListTopics(list),
		Start:   NewStartTask(platform),
		History: NewGetHistory(platform),
		Note:    NewSaveNote(platform),
	}
}
