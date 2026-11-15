package tasks

import (
	"strings"
	"time"
	"unicode/utf8"
)

const MaxNote = 500

type EntryKind string

const (
	AttemptEntry EntryKind = "attempt"
	NoteEntry    EntryKind = "note"
)

// Entry is one line of a student's record: an attempt the grader wrote, or a note on a task.
type Entry struct {
	Kind   EntryKind
	TaskID string
	At     time.Time
	Earned int
	Total  int
	Failed []string
	Text   string
}

func NewAttempt(taskID string, checked time.Time, grade Grade) Entry {
	return Entry{
		Kind: AttemptEntry, TaskID: taskID, At: checked, Earned: grade.Earned, Total: grade.Total, Failed: grade.Failed(),
	}
}

func NewNote(taskID string, written time.Time, text string) (Entry, error) {
	text = strings.TrimSpace(text)

	switch {
	case text == "":
		return Entry{}, ErrEmptyNote
	case utf8.RuneCountInString(text) > MaxNote:
		return Entry{}, ErrNoteTooLong
	}

	return Entry{Kind: NoteEntry, TaskID: taskID, At: written, Text: text}, nil
}
