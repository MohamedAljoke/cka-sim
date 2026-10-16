package domainerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

var errNotFound = domainerr.New("task_not_found", "no task has that id")

func TestAWrappedErrorStillMatchesItsSentinel(t *testing.T) {
	err := fmt.Errorf("%w: %w", errNotFound, errors.New("disk gone"))

	var found *domainerr.Error
	if !errors.Is(err, errNotFound) || !errors.As(err, &found) || found.Code != "task_not_found" {
		t.Fatalf("err = %v, found = %+v", err, found)
	}
}

func TestErrorsWithTheSameCodeAreStillDifferentSentinels(t *testing.T) {
	other := domainerr.New("task_not_found", "no task has that id")

	if errors.Is(other, errNotFound) {
		t.Fatal("a second error with the same code matched the first sentinel")
	}
}

func TestTheTextIsTheCodeAndMessage(t *testing.T) {
	if got := errNotFound.Error(); got != "task_not_found: no task has that id" {
		t.Fatalf("Error() = %q", got)
	}
}
