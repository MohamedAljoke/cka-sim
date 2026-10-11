package grader

import (
	"reflect"
	"testing"
)

func TestParseAddsUpPassAndFail(t *testing.T) {
	got := Parse("PASS 2 Deployment web wants 4 replicas\nFAIL 3 4 Pods of web are Ready\n")

	want := Result{
		Checks: []Check{
			{Passed: true, Points: 2, Description: "Deployment web wants 4 replicas"},
			{Passed: false, Points: 3, Description: "4 Pods of web are Ready"},
		},
		Earned: 2,
		Total:  5,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

func TestParseIgnoresOtherLines(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{"kubectl noise", "deployment.apps/web scaled"},
		{"lowercase verb", "pass 2 web scaled"},
		{"points not a number", "PASS x web scaled"},
		{"zero points", "PASS 0 web scaled"},
		{"negative points", "FAIL -1 web scaled"},
		{"no description", "PASS 2"},
		{"blank description", "PASS 2 "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.line + "\nPASS 1 kept\n")

			if len(got.Checks) != 1 || got.Checks[0].Description != "kept" || got.Total != 1 {
				t.Errorf("got %+v, want only the kept check", got)
			}
		})
	}
}

func TestParseEmptyOutput(t *testing.T) {
	got := Parse("")

	if got.Checks == nil || len(got.Checks) != 0 || got.Earned != 0 || got.Total != 0 {
		t.Errorf("got %+v, want 0/0 with empty checks", got)
	}
}
