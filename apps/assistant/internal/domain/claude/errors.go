package claude

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrBadStudent = domainerr.New("bad_student",
		"Claude Code runs for a student, named with 1 to 30 lowercase letters, digits or dashes")
	ErrNotInstalled = domainerr.New("claude_not_installed",
		"Claude Code isn't installed where the assistant runs; install it there, or set CKA_SIM_CLAUDE")
)
