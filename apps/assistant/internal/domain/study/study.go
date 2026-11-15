package study

import "time"

// Filter keeps the tasks in any of Domains, with any of Topics, whose id or title has every word of Search.
type Filter struct {
	Domains []string
	Topics  []string
	Search  string
}

// TaskSummary is a catalog task without its question, so a list never gives a task away.
type TaskSummary struct {
	ID     string
	Title  string
	Domain string
	Topics []string
	Weight int
}

type Task struct {
	ID       string
	Title    string
	Domain   string
	Host     string
	Question string
}

// Started is a task set up on the student's cluster. When it isn't Ready, ExitCode and Output say why.
type Started struct {
	Task     Task
	Ready    bool
	ExitCode int
	Output   string
}

type EntryKind string

const (
	AttemptEntry EntryKind = "attempt"
	NoteEntry    EntryKind = "note"
)

// Entry is an attempt the grader recorded, with the checks that failed, or a note on a task.
type Entry struct {
	Kind   EntryKind
	TaskID string
	At     time.Time
	Earned int
	Total  int
	Failed []string
	Text   string
}

type Domain struct {
	Name   string
	Tasks  int
	Topics []Topic
}

type Topic struct {
	Name  string
	Tasks int
}
