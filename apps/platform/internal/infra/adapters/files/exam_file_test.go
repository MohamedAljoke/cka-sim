package files

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestAnExamSavedIsTheExamLoaded(t *testing.T) {
	f := NewExamFile(filepath.Join(t.TempDir(), "state", "exam.json"))
	start := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	exam := exams.Exam{Minutes: 30, Started: start, Deadline: start.Add(30 * time.Minute), Tasks: []exams.Task{
		{ID: "tr-etcd", Weight: 7, Last: true, Setup: exams.SetUp, Flagged: true,
			Grade: &tasks.Grade{Earned: 1, Total: 2, Checks: []tasks.Check{{Passed: true, Points: 1, Description: "logged"}}}},
		{ID: "nw-svc", Weight: 3, Setup: exams.SetupFailed, SetupError: "exit code 1"},
	}}

	if err := f.Save(exam); err != nil {
		t.Fatal(err)
	}
	got, err := f.Load()

	if err != nil || got == nil || !reflect.DeepEqual(*got, exam) {
		t.Errorf("Load = %+v, %v", got, err)
	}
}

func TestNoFileIsNoExam(t *testing.T) {
	got, err := NewExamFile(filepath.Join(t.TempDir(), "exam.json")).Load()

	if got != nil || err != nil {
		t.Errorf("Load = %+v, %v", got, err)
	}
}

func TestAClearedExamIsGone(t *testing.T) {
	f := NewExamFile(filepath.Join(t.TempDir(), "exam.json"))
	f.Save(exams.Exam{Minutes: 30})

	if err := f.Clear(); err != nil {
		t.Fatal(err)
	}
	if got, err := f.Load(); got != nil || err != nil {
		t.Errorf("Load = %+v, %v", got, err)
	}
	if err := f.Clear(); err != nil {
		t.Errorf("clearing twice: %v", err)
	}
}

func TestABrokenFileIsAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "exam.json")
	os.WriteFile(path, []byte("{half"), 0o600)

	if _, err := NewExamFile(path).Load(); err == nil {
		t.Error("Load read a broken file")
	}
}
