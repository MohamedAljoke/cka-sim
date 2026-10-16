package domainerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

var errUnavailable = domainerr.Conflict(
	"lab_unavailable", "the lab is not reachable right now").AsFault()

func TestDomainErrorsExposeTheirCodeToAReporter(t *testing.T) {
	var coded interface{ ErrorCode() string }

	if !errors.As(fmt.Errorf("reach the lab: %w", errUnavailable), &coded) {
		t.Fatal("a wrapped domain error does not satisfy the reporter's interface")
	}
	if coded.ErrorCode() != "lab_unavailable" {
		t.Fatalf("code = %q", coded.ErrorCode())
	}
}

func TestDetailsDoNotChangeTheCode(t *testing.T) {
	located := errUnavailable.With(map[string]any{"attempt": 3})

	if located.ErrorCode() != errUnavailable.ErrorCode() {
		t.Fatalf("code = %q, want %q", located.ErrorCode(), errUnavailable.ErrorCode())
	}
}
