package exams

import (
	"errors"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestNewPreparesEveryDrawnTask(t *testing.T) {
	e := New([]tasks.Task{{ID: "tr-etcd", Weight: 7, Order: tasks.Last}, {ID: "wl-scale", Weight: 4}}, 120)

	want := []Task{{ID: "tr-etcd", Weight: 7, Last: true, Setup: Preparing}, {ID: "wl-scale", Weight: 4, Setup: Preparing}}
	if e.Minutes != 120 || len(e.Tasks) != 2 || e.Tasks[0] != want[0] || e.Tasks[1] != want[1] || !e.SettingUp() {
		t.Errorf("New = %+v", e)
	}
}

func TestValidateKeepsCountAndMinutesInRange(t *testing.T) {
	for _, tc := range []struct {
		count, minutes int
		ok             bool
	}{{1, 1, true}, {16, 240, true}, {0, 120, false}, {16, 0, false}, {16, 241, false}} {
		if err := Validate(tc.count, tc.minutes); (err == nil) != tc.ok || (err != nil && !errors.Is(err, ErrInvalid)) {
			t.Errorf("Validate(%d, %d) = %v", tc.count, tc.minutes, err)
		}
	}
}

func TestAnExamExpiresAtItsDeadlineOnlyWhileRunning(t *testing.T) {
	start := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	e := Exam{Started: start, Deadline: start.Add(time.Hour)}

	if e.Expired(start.Add(59*time.Minute)) || !e.Expired(start.Add(time.Hour)) {
		t.Error("a running exam expires exactly at its deadline")
	}

	e.Ended = start.Add(10 * time.Minute)
	if e.Expired(start.Add(2 * time.Hour)) {
		t.Error("an ended exam doesn't expire")
	}

	if (Exam{}).Expired(start) {
		t.Error("an exam still setting up doesn't expire")
	}
}

func TestAnEndedExamScoresUntilItIsScored(t *testing.T) {
	e := Exam{Ended: time.Now()}
	if !e.Over() || !e.Scoring() {
		t.Errorf("ended: over %v, scoring %v", e.Over(), e.Scoring())
	}

	e.Scored = time.Now()
	if e.Scoring() {
		t.Error("a scored exam isn't scoring")
	}
}
