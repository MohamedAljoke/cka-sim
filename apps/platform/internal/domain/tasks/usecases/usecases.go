package usecases

import "time"

type UseCases struct {
	List        *ListTasks
	Get         *GetTask
	Current     *GetCurrentTask
	Solution    *GetSolution
	Heal        *HealCluster
	SetUp       *SetUpTask
	Start       *StartTask
	MoveTo      *MoveToTask
	Grade       *GradeTask
	Check       *CheckTask
	Tidy        *TidyTask
	TidyCurrent *TidyCurrentTask
	Solve       *SolveTask
	Selftest    *SelftestTask
	History     *GetHistory
	Note        *SaveNote
}

func New(repo TaskRepository, runner ScriptRunner, attempts AttemptRepository, now func() time.Time) UseCases {
	clusters := newClusters()
	heal := NewHealCluster(repo, runner, clusters)
	setUp := NewSetUpTask(repo, runner, clusters)
	start := NewStartTask(repo, heal, setUp)
	grade := NewGradeTask(repo, runner)
	tidy := NewTidyTask(repo, runner, clusters)
	solve := NewSolveTask(repo, runner, clusters)

	return UseCases{
		List:        NewListTasks(repo),
		Get:         NewGetTask(repo),
		Current:     NewGetCurrentTask(repo, clusters),
		Solution:    NewGetSolution(repo),
		Heal:        heal,
		SetUp:       setUp,
		Start:       start,
		MoveTo:      NewMoveToTask(repo, tidy, start, clusters),
		Grade:       grade,
		Check:       NewCheckTask(grade, attempts, now),
		Tidy:        tidy,
		TidyCurrent: NewTidyCurrentTask(tidy, clusters),
		Solve:       solve,
		Selftest:    NewSelftestTask(start, grade, solve),
		History:     NewGetHistory(attempts),
		Note:        NewSaveNote(repo, attempts, now),
	}
}
