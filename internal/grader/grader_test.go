package grader

import "testing"

func TestParseScoresPassAndFailLines(t *testing.T) {
	out := `waiting for node…
PASS 4 node is Ready
FAIL 2 kubelet is enabled
some diagnostic
PASS x not a number
PASS 1 file has the answer`
	r := Parse("t", out)
	if len(r.Checks) != 3 {
		t.Fatalf("got %d checks, want 3: %+v", len(r.Checks), r.Checks)
	}
	if r.Earned != 5 || r.Total != 7 {
		t.Fatalf("got %d/%d, want 5/7", r.Earned, r.Total)
	}
	if got := r.Fraction(); got < 0.714 || got > 0.715 {
		t.Fatalf("fraction %v", got)
	}
}

func TestParseEmptyOutputScoresZero(t *testing.T) {
	if r := Parse("t", ""); r.Total != 0 || r.Fraction() != 0 {
		t.Fatalf("unexpected %+v", r)
	}
}
