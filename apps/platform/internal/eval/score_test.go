package eval

import (
	"bytes"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/grader"
)

var (
	partial = grader.Result{Earned: 2, Total: 4}
	full    = grader.Result{Earned: 4, Total: 4}
	nothing = grader.Result{Earned: 0, Total: 4}
)

func TestScoreAGoodReview(t *testing.T) {
	a := Attempt{Name: "half", Mentions: []string{"targetPort"}}
	r := Review{Tools: []string{"current_task", "run_check"}, Verdict: &Verdict{Passed: false, Mistakes: []string{"the TargetPort is 8080"}}}

	got := Score(a, partial, r)

	if want := (Marks{Verdict: Pass, Checked: Pass, NoSpoiler: Pass, Named: Pass}); got != want {
		t.Errorf("marks = %+v, want %+v", got, want)
	}
}

func TestScoreABadReview(t *testing.T) {
	a := Attempt{Name: "half", Mentions: []string{"targetPort"}}
	r := Review{Tools: []string{"explain"}, Verdict: &Verdict{Passed: true, Summary: "All good, the selector is right."}}

	got := Score(a, partial, r)

	if want := (Marks{Verdict: Fail, Checked: Fail, NoSpoiler: Fail, Named: Fail}); got != want {
		t.Errorf("marks = %+v, want %+v", got, want)
	}
}

func TestScoreNeedsEveryMention(t *testing.T) {
	a := Attempt{Name: Untouched, Mentions: []string{"selector", "targetPort"}}
	r := Review{Verdict: &Verdict{Hint: "check the selector"}}

	if got := Score(a, nothing, r).Named; got != Fail {
		t.Errorf("named = %s, want fail with targetPort missing", got)
	}
}

func TestScoreSkipsNamedWithoutMentions(t *testing.T) {
	r := Review{Tools: []string{"run_check"}, Verdict: &Verdict{Passed: true}}

	marks := Score(Attempt{Name: Solved}, full, r)
	row := Row{Marks: marks}

	if marks.Named != None {
		t.Errorf("named = %s, want -", marks.Named)
	}
	if got, of := row.Score(); got != 3 || of != 3 {
		t.Errorf("score %d/%d, want 3/3", got, of)
	}
}

func TestScoreAReviewWithNoVerdict(t *testing.T) {
	marks := Score(Attempt{Name: Solved}, full, Review{Tools: []string{"run_check"}, Failure: "error_max_turns"})

	if marks.Verdict != Fail || marks.Checked != Pass {
		t.Errorf("marks = %+v, want the verdict failed", marks)
	}
}

func TestAnAttemptMustBeWhatItSays(t *testing.T) {
	tests := []struct {
		attempt string
		truth   grader.Result
		valid   bool
	}{
		{Solved, full, true},
		{Solved, partial, false},
		{Untouched, nothing, true},
		{Untouched, partial, false},
		{"half", partial, true},
		{"half", full, false},
		{"half", nothing, false},
		{Solved, grader.Result{}, false},
	}
	for _, tt := range tests {
		if got := invalid(Attempt{Name: tt.attempt}, tt.truth) == ""; got != tt.valid {
			t.Errorf("%s earning %d/%d: valid %v, want %v", tt.attempt, tt.truth.Earned, tt.truth.Total, got, tt.valid)
		}
	}
}

func TestTableTotalsOnlyScoredMarks(t *testing.T) {
	rows := []Row{
		{Task: "wl-scale", Attempt: Solved, Truth: full, Marks: Marks{Verdict: Pass, Checked: Pass, NoSpoiler: Pass, Named: None}},
		{Task: "wl-scale", Attempt: "half", Truth: full, Marks: unscored, Invalid: "a partial attempt must earn some points"},
	}
	var out bytes.Buffer

	if err := Table(&out, rows); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "score 3/3") || !strings.Contains(out.String(), "invalid attempt") {
		t.Errorf("table:\n%s", out.String())
	}
}
