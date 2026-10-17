package labs

type ScriptResult struct {
	Output   string
	ExitCode int
}

func (r ScriptResult) Passed() bool { return r.ExitCode == 0 }
