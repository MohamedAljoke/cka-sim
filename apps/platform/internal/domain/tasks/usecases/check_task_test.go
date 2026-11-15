package usecases

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

var checkedAt = time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)

func TestAPracticeCheckSavesTheAttemptWithWhatFailed(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "PASS 2 web has 4 replicas\nFAIL 1 web is Ready\n"}}
	attempts := &fakeAttempts{}

	got, err := newCheck(r, attempts).Execute(context.Background(), owner, "wl-scale")

	if err != nil || got.Grade.Earned != 2 || len(attempts.entries) != 1 {
		t.Fatalf("Execute = %+v, %v; saved %+v", got, err, attempts.entries)
	}
	a := attempts.entries[0]
	if a.Kind != tasks.AttemptEntry || a.TaskID != "wl-scale" || !a.At.Equal(checkedAt) || a.Earned != 2 || a.Total != 3 ||
		!slices.Equal(a.Failed, []string{"web is Ready"}) {
		t.Errorf("saved %+v", a)
	}
}

func TestEveryCheckAddsAnAttempt(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "PASS 1 done\n"}}
	attempts := &fakeAttempts{}
	check := newCheck(r, attempts)

	check.Execute(context.Background(), owner, "wl-scale")
	check.Execute(context.Background(), owner, "wl-scale")

	if len(attempts.entries) != 2 {
		t.Errorf("saved %+v", attempts.entries)
	}
}

func TestACheckThatPrintsNoResultSavesNoAttempt(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "bash: jq: command not found\n", ExitCode: 127}}
	attempts := &fakeAttempts{}

	got, err := newCheck(r, attempts).Execute(context.Background(), owner, "wl-scale")

	if err != nil || got.Script.ExitCode != 127 || len(attempts.entries) != 0 {
		t.Errorf("Execute = %+v, %v; saved %+v", got, err, attempts.entries)
	}
}

func TestACheckOnALabThatIsNotReadySavesNoAttempt(t *testing.T) {
	attempts := &fakeAttempts{}

	_, err := newCheck(&fakeRunner{err: labs.ErrNotReady}, attempts).Execute(context.Background(), owner, "wl-scale")

	if !errors.Is(err, labs.ErrNotReady) || len(attempts.entries) != 0 {
		t.Errorf("got %v; saved %+v", err, attempts.entries)
	}
}

func TestACheckWhoseAttemptCannotBeSavedStillAnswersItsGrade(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "PASS 1 done\n"}}

	got, err := newCheck(r, &fakeAttempts{err: errors.New("disk full")}).Execute(context.Background(), owner, "wl-scale")

	if err != nil || got.Grade.Earned != 1 {
		t.Errorf("Execute = %+v, %v", got, err)
	}
}

func newCheck(r ScriptRunner, attempts AttemptRepository) *CheckTask {
	repo := fakeRepository{tasks: []tasks.Task{withCheck}}
	return NewCheckTask(NewGradeTask(repo, r), attempts, func() time.Time { return checkedAt })
}
