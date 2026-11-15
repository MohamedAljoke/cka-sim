package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestANoteIsSavedOnItsTask(t *testing.T) {
	attempts := &fakeAttempts{}

	err := newSaveNote(attempts).Execute(context.Background(), owner, "wl-scale", "  mixed up ingress and egress  ")

	if err != nil || len(attempts.entries) != 1 {
		t.Fatalf("got %v; saved %+v", err, attempts.entries)
	}
	n := attempts.entries[0]
	if n.Kind != tasks.NoteEntry || n.TaskID != "wl-scale" || n.Text != "mixed up ingress and egress" || !n.At.Equal(checkedAt) {
		t.Errorf("saved %+v", n)
	}
}

func TestANoteOver500CharactersIsRefused(t *testing.T) {
	attempts := &fakeAttempts{}

	err := newSaveNote(attempts).Execute(context.Background(), owner, "wl-scale", strings.Repeat("é", 501))

	if !errors.Is(err, tasks.ErrNoteTooLong) || len(attempts.entries) != 0 {
		t.Errorf("got %v; saved %+v", err, attempts.entries)
	}
}

func TestAnEmptyNoteIsRefused(t *testing.T) {
	err := newSaveNote(&fakeAttempts{}).Execute(context.Background(), owner, "wl-scale", " \n ")

	if !errors.Is(err, tasks.ErrEmptyNote) {
		t.Errorf("got %v", err)
	}
}

func TestANoteOnAnUnknownTaskIsNotFound(t *testing.T) {
	attempts := &fakeAttempts{}

	err := newSaveNote(attempts).Execute(context.Background(), owner, "no-such-task", "a note")

	if !errors.Is(err, tasks.ErrNotFound) || len(attempts.entries) != 0 {
		t.Errorf("got %v; saved %+v", err, attempts.entries)
	}
}

func TestANoteThatCannotBeSavedIsUnavailableAndKeepsTheCause(t *testing.T) {
	cause := errors.New("disk full")

	err := newSaveNote(&fakeAttempts{err: cause}).Execute(context.Background(), owner, "wl-scale", "a note")

	if !errors.Is(err, tasks.ErrNoHistory) || !errors.Is(err, cause) {
		t.Errorf("got %v", err)
	}
}

func newSaveNote(attempts AttemptRepository) *SaveNote {
	return NewSaveNote(fakeRepository{tasks: []tasks.Task{withCheck}}, attempts, func() time.Time { return checkedAt })
}
