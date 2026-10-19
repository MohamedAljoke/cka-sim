package tutor

// Task is what the student is asked to do, as the platform tells it.
type Task struct {
	ID       string
	Title    string
	Domain   string
	Host     string
	Question string
}

type Check struct {
	Passed      bool
	Points      int
	Description string
}

// Grade is the grader's verdict. When it ran no checks, ExitCode and Output say why.
type Grade struct {
	TaskID   string
	Earned   int
	Total    int
	Checks   []Check
	ExitCode int
	Output   string
}

type Inspection struct {
	Output    string
	ExitCode  int
	Truncated bool
}
