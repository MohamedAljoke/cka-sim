package kind

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	kindlog "sigs.k8s.io/kind/pkg/log"
)

// kind reports each phase of a build as a " • <step>  ..." line to a logger that has no spinner.
type stepLogger struct {
	progress func(step string)
}

func (l stepLogger) Warn(message string)               { log.Print(message) }
func (l stepLogger) Warnf(format string, args ...any)  { log.Printf(format, args...) }
func (l stepLogger) Error(message string)              { log.Print(message) }
func (l stepLogger) Errorf(format string, args ...any) { log.Printf(format, args...) }

func (l stepLogger) V(level kindlog.Level) kindlog.InfoLogger {
	return infoLogger{enabled: level == 0, progress: l.progress}
}

type infoLogger struct {
	enabled  bool
	progress func(step string)
}

func (l infoLogger) Info(message string) {
	if !l.enabled {
		return
	}

	log.Print(strings.TrimSpace(message))

	if step, ok := stepOf(message); ok {
		l.progress(step)
	}
}

func (l infoLogger) Infof(format string, args ...any) { l.Info(fmt.Sprintf(format, args...)) }

func (l infoLogger) Enabled() bool { return l.enabled }

func stepOf(line string) (string, bool) {
	step, ok := strings.CutPrefix(strings.TrimSpace(line), "• ")
	if !ok {
		return "", false
	}
	step = strings.TrimSuffix(step, "...")

	return strings.TrimRightFunc(step, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != ')'
	}), true
}
