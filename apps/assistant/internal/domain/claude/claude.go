package claude

import "regexp"

type TerminalSize struct {
	Cols uint16
	Rows uint16
}

// A student's name becomes the folder of their Claude Code login, so it stays a plain, short name.
var studentName = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,28}[a-z0-9])?$`)

func ValidStudent(name string) bool {
	return studentName.MatchString(name)
}
