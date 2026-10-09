package main

import (
	"bytes"
	"context"
	"slices"
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

func TestShellArgs(t *testing.T) {
	tests := []struct {
		name string
		tty  bool
		want []string
	}{
		{name: "terminal", tty: true, want: []string{"exec", "-i", "-t", "-u", "candidate", "-w", "/home/candidate", "cka-sim-base", "bash", "-l"}},
		{name: "pipe", tty: false, want: []string{"exec", "-i", "-u", "candidate", "-w", "/home/candidate", "cka-sim-base", "bash", "-l"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shellArgs(tt.tty); !slices.Equal(got, tt.want) {
				t.Errorf("shellArgs(%v) = %q, want %q", tt.tty, got, tt.want)
			}
		})
	}
}
