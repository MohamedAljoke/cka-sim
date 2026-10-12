package eval

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

func TestLoadReadsEachTasksAttempts(t *testing.T) {
	fsys := fstest.MapFS{
		"wl-scale/evals.yaml": {Data: []byte("- name: solved\n- name: half\n  mentions: [replicas]\n  script: kubectl scale\n")},
	}

	set, err := Load(fsys, []tasks.Task{{ID: "wl-scale", Dir: "wl-scale"}, {ID: "tr-dns", Dir: "tr-dns"}})

	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 2 || set[0].Name != Solved || set[1].Name != "half" || set[1].Mentions[0] != "replicas" {
		t.Fatalf("set = %+v", set)
	}
	if set[1].Task.ID != "wl-scale" || set[1].Script != "kubectl scale" {
		t.Errorf("half = %+v, want it tied to wl-scale with its script", set[1])
	}
}

func TestLoadRefusesABadSet(t *testing.T) {
	tests := []struct{ name, yaml, wantErr string }{
		{"custom without a script", "- name: half\n", "needs a script"},
		{"built in with a script", "- name: solved\n  script: echo\n", "takes no script"},
		{"listed twice", "- name: solved\n- name: solved\n", "listed twice"},
		{"no name", "- mentions: [x]\n", "no name"},
		{"unknown field", "- name: solved\n  expect: pass\n", "unknown field"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{"x/evals.yaml": {Data: []byte(tt.yaml)}}

			_, err := Load(fsys, []tasks.Task{{ID: "x", Dir: "x"}})

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestOnlyKeepsOneTask(t *testing.T) {
	set := []Attempt{{Task: tasks.Task{ID: "a"}, Name: Solved}, {Task: tasks.Task{ID: "b"}, Name: Solved}}

	if got := Only(set, "b"); len(got) != 1 || got[0].Task.ID != "b" {
		t.Errorf("Only = %+v", got)
	}
	if len(set) != 2 {
		t.Errorf("Only changed the set it was given: %+v", set)
	}
}
