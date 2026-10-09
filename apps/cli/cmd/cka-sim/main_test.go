package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		args       []string
		wantOutput string
		wantErr    string
	}{
		{args: nil, wantOutput: "cka-sim doctor"},
		{args: []string{"help"}, wantOutput: "cka-sim doctor"},
		{args: []string{"version"}, wantOutput: "cka-sim "},
		{args: []string{"nope"}, wantErr: `unknown command "nope"`},
		{args: []string{"version", "extra"}, wantErr: "takes no arguments"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			var out bytes.Buffer

			err := run(context.Background(), tt.args, &out)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("got error %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(out.String(), tt.wantOutput) {
				t.Errorf("output %q does not contain %q", out.String(), tt.wantOutput)
			}
		})
	}
}
