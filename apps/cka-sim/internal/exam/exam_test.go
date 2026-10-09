package exam

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/grader"
)

func TestScoreWeighsEachTask(t *testing.T) {
	e := Exam{Tasks: []Task{
		{ID: "tr", Result: &grader.Result{Earned: 4, Total: 4}},
		{ID: "wl", Result: &grader.Result{Earned: 0, Total: 4}},
	}}

	percent, passed := Score(e, map[string]int{"tr": 6, "wl": 4})

	if percent != 60 || passed {
		t.Errorf("got %d%% passed=%v, want 60%% and a fail", percent, passed)
	}
}

func TestScorePassesAtSixtySix(t *testing.T) {
	e := Exam{Tasks: []Task{
		{ID: "a", Result: &grader.Result{Earned: 2, Total: 3}},
		{ID: "b", Result: &grader.Result{Earned: 2, Total: 3}},
	}}

	percent, passed := Score(e, map[string]int{"a": 5, "b": 5})

	if percent != 67 || !passed {
		t.Errorf("got %d%% passed=%v, want 67%% and a pass", percent, passed)
	}
}

func TestScoreCountsUngradedTasksAsZero(t *testing.T) {
	e := Exam{Tasks: []Task{
		{ID: "a", Result: &grader.Result{Earned: 4, Total: 4}},
		{ID: "b"},
	}}

	if percent, _ := Score(e, map[string]int{"a": 1, "b": 1}); percent != 50 {
		t.Errorf("got %d%%, want 50%%", percent)
	}
}

func TestFileStoreRoundTrip(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "state", "exam.json")}
	saved := Exam{
		Tasks:    []Task{{ID: "wl-scale", Setup: Ready, Flagged: true}},
		Minutes:  30,
		Started:  time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC),
		Deadline: time.Date(2026, 10, 9, 12, 30, 0, 0, time.UTC),
	}

	if err := store.Save(saved); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded == nil || !reflect.DeepEqual(*loaded, saved) {
		t.Errorf("loaded %+v, want %+v", loaded, saved)
	}
}

func TestFileStoreWithoutAFileHasNoExam(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "exam.json")}

	e, err := store.Load()

	if err != nil || e != nil {
		t.Errorf("got %+v, %v, want no exam and no error", e, err)
	}
}

func TestFileStoreClear(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "exam.json")}
	if err := store.Save(Exam{Minutes: 5}); err != nil {
		t.Fatal(err)
	}

	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}

	if e, _ := store.Load(); e != nil {
		t.Errorf("after Clear, Load gave %+v", e)
	}
	if err := store.Clear(); err != nil {
		t.Errorf("clearing twice: %v", err)
	}
}
