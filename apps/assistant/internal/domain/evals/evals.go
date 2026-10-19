package evals

// Question is what a stuck student asks, once with the task broken and once after its solution.
const Question = "My check fails, why?"

type Check struct {
	Passed      bool
	Points      int
	Description string
}

func Failing(checks []Check) []Check {
	var failing []Check

	for _, c := range checks {
		if !c.Passed {
			failing = append(failing, c)
		}
	}

	return failing
}

// Script is how one of the task's scripts (setup, solution) ran on the lab.
type Script struct {
	Passed   bool
	ExitCode int
	Output   string
}

type Answer struct {
	Text  string
	Tools []string
}

// Diagnosis is the judge's verdict on the answer to the broken task.
type Diagnosis struct {
	// Named are the failing checks' descriptions the answer points at.
	Named    []string
	GaveAway bool
	// Quote is the part of the answer that gave the solution away.
	Quote string
	Why   string
}

// Recheck is the judge's verdict on the answer after the solution, when every check passes.
type Recheck struct {
	NothingWrong bool
	Why          string
}

// Result is one task's eval. Diagnosis and Recheck stay nil when the eval stopped before them, and Err says why.
type Result struct {
	TaskID    string
	Failing   []Check
	Broken    Answer
	Diagnosis *Diagnosis
	Solved    Answer
	Recheck   *Recheck
	Err       error
}

func (r Result) Passed() bool {
	return r.Err == nil && r.Diagnosis != nil && len(r.Diagnosis.Named) == len(r.Failing) &&
		!r.Diagnosis.GaveAway && r.Recheck != nil && r.Recheck.NothingWrong
}
