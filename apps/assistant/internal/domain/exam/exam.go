package exam

import "time"

type Selection struct {
	Domains []string
	Topics  []string
	TaskIDs []string
}

type Task struct {
	ID    string
	Title string
}

type Exam struct {
	Tasks   []Task
	Minutes int
}

type TaskResult struct {
	ID          string
	Title       string
	Weight      int
	Earned      int
	Total       int
	Failed      []string
	Flagged     bool
	SetupFailed bool
	Error       string
}

type Result struct {
	Tasks       []TaskResult
	Minutes     int
	Started     time.Time
	Ended       time.Time
	Percent     int
	Passed      bool
	PassPercent int
}
