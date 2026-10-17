package exams

import (
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestScoreWeighsEachTasksShareByItsWeight(t *testing.T) {
	e := Exam{Tasks: []Task{graded(6, 3, 3), graded(4, 1, 2)}}

	percent, passed := Score(e)

	if percent != 80 || !passed {
		t.Errorf("Score = %d, %v; want 80, true", percent, passed)
	}
}

func TestScorePassesAtSixtySixPercent(t *testing.T) {
	for _, tc := range []struct {
		earned  int
		percent int
		passed  bool
	}{{66, 66, true}, {65, 65, false}} {
		percent, passed := Score(Exam{Tasks: []Task{graded(1, tc.earned, 100)}})

		if percent != tc.percent || passed != tc.passed {
			t.Errorf("%d/100: Score = %d, %v", tc.earned, percent, passed)
		}
	}
}

func TestATaskWithoutAGradeOrPointsScoresZero(t *testing.T) {
	e := Exam{Tasks: []Task{graded(5, 5, 5), {ID: "b", Weight: 5, Setup: SetUp}, graded(5, 0, 0)}}

	if percent, _ := Score(e); percent != 33 {
		t.Errorf("Score = %d, want 33", percent)
	}
}

func TestATaskWhoseSetupFailedIsLeftOutOfTheScore(t *testing.T) {
	e := Exam{Tasks: []Task{graded(5, 5, 5), {ID: "b", Weight: 5, Setup: SetupFailed}}}

	if percent, passed := Score(e); percent != 100 || !passed {
		t.Errorf("Score = %d, %v; want 100, true", percent, passed)
	}
}

func TestAnExamWithNothingPlayableScoresZero(t *testing.T) {
	e := Exam{Tasks: []Task{{ID: "a", Weight: 5, Setup: SetupFailed}}}

	if percent, passed := Score(e); percent != 0 || passed {
		t.Errorf("Score = %d, %v", percent, passed)
	}
}

func graded(weight, earned, total int) Task {
	return Task{ID: "t", Weight: weight, Setup: SetUp, Grade: &tasks.Grade{Earned: earned, Total: total}}
}
