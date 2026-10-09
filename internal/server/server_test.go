package server

import (
	"testing"

	"github.com/MohamedAljoke/cka-sim/internal/grader"
	"github.com/MohamedAljoke/cka-sim/internal/tasks"
)

func TestScoreWeighsPartialCredit(t *testing.T) {
	exam := []tasks.Task{{ID: "a", Weight: 8}, {ID: "b", Weight: 4}, {ID: "c", Weight: 8}}
	results := []grader.Result{
		{TaskID: "a", Earned: 4, Total: 4}, // 8 of 8
		{TaskID: "b", Earned: 1, Total: 2}, // 2 of 4
		// c was never graded: 0 of 8
	}
	if got := Score(exam, results); got != 50 {
		t.Fatalf("score %v, want 50", got)
	}
}
