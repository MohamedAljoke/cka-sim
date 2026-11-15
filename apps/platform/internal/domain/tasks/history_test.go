package tasks

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestAnAttemptKeepsTheChecksThatFailed(t *testing.T) {
	grade := Grade{Earned: 2, Total: 4, Checks: []Check{
		{Passed: true, Points: 2, Description: "scaled"},
		{Passed: false, Points: 1, Description: "labelled"},
		{Passed: false, Points: 1, Description: "Ready"},
	}}

	got := NewAttempt("wl-scale", time.Unix(0, 0), grade)

	if got.Kind != AttemptEntry || got.Earned != 2 || got.Total != 4 || !slices.Equal(got.Failed, []string{"labelled", "Ready"}) {
		t.Errorf("got %+v", got)
	}
}

func TestANoteIsAtMost500Characters(t *testing.T) {
	if _, err := NewNote("wl-scale", time.Unix(0, 0), strings.Repeat("ü", 500)); err != nil {
		t.Errorf("500 characters: %v", err)
	}
	if _, err := NewNote("wl-scale", time.Unix(0, 0), strings.Repeat("ü", 501)); !errors.Is(err, ErrNoteTooLong) {
		t.Errorf("501 characters: %v", err)
	}
}

func TestANoteOfOnlySpacesIsEmpty(t *testing.T) {
	if _, err := NewNote("wl-scale", time.Unix(0, 0), "  \t\n"); !errors.Is(err, ErrEmptyNote) {
		t.Errorf("got %v", err)
	}
}
