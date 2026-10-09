package tasks

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParse(t *testing.T) {
	task, err := Parse(taskMD("wl-scale", "domain: workloads\ntopics: [deployments, scaling]\nweight: 4") + "\n## Task\n\nScale it.\n")
	if err != nil {
		t.Fatal(err)
	}

	want := Task{ID: "wl-scale", Title: "A task", Domain: Workloads, Topics: []string{"deployments", "scaling"}, Weight: 4, Question: "## Task\n\nScale it."}
	if task.ID != want.ID || task.Title != want.Title || task.Domain != want.Domain ||
		!slices.Equal(task.Topics, want.Topics) || task.Weight != want.Weight || task.Question != want.Question {
		t.Errorf("got %+v\nwant %+v", task, want)
	}
}

func TestParseRejects(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr string
	}{
		{"no frontmatter", "## Task\n", "missing frontmatter"},
		{"unclosed frontmatter", "---\nid: x\n", "no closing ---"},
		{"unknown key", taskMD("x", "domain: workloads\nweight: 1\ncluster: a"), "unknown field"},
		{"no id", taskMD("", "domain: workloads\nweight: 1"), "id is required"},
		{"no title", "---\nid: x\ndomain: workloads\nweight: 1\n---\n", "title is required"},
		{"unknown domain", taskMD("x", "domain: security\nweight: 1"), `unknown domain "security"`},
		{"no weight", taskMD("x", "domain: workloads"), "weight must be more than 0"},
		{"capital topic", taskMD("x", "domain: workloads\nweight: 1\ntopics: [Deployments]"), "lowercase-kebab"},
		{"spaced topic", taskMD("x", "domain: workloads\nweight: 1\ntopics: [rolling update]"), "lowercase-kebab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.text)

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got error %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadSortsByID(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	addTask(fsys, "net-svc")

	all, err := Load(fsys)
	if err != nil {
		t.Fatal(err)
	}

	if ids := []string{all[0].ID, all[1].ID}; !slices.Equal(ids, []string{"net-svc", "wl-scale"}) {
		t.Errorf("ids = %q, want sorted", ids)
	}
	if all[0].Dir != "net-svc" || all[0].Explain != "Why." {
		t.Errorf("dir/explain = %q/%q", all[0].Dir, all[0].Explain)
	}
}

func TestLoadRejects(t *testing.T) {
	tests := []struct {
		name    string
		breakIt func(fstest.MapFS)
		wantErr string
	}{
		{"missing check.sh", func(f fstest.MapFS) { delete(f, "wl-scale/check.sh") }, "wl-scale/task.md: missing check.sh"},
		{"missing explain.md", func(f fstest.MapFS) { delete(f, "wl-scale/explain.md") }, "missing explain.md"},
		{"id differs from folder", func(f fstest.MapFS) {
			f["wl-scale/task.md"] = &fstest.MapFile{Data: []byte(taskMD("wl-other", "domain: workloads\nweight: 1"))}
		}, `id "wl-other" must match its folder "wl-scale"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{}
			addTask(fsys, "wl-scale")
			tt.breakIt(fsys)

			_, err := Load(fsys)

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got error %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestFind(t *testing.T) {
	all := []Task{{ID: "a"}, {ID: "b"}}

	if got, ok := Find(all, "b"); !ok || got.ID != "b" {
		t.Errorf("Find(b) = %+v, %v", got, ok)
	}
	if _, ok := Find(all, "c"); ok {
		t.Error("Find(c) found a task that doesn't exist")
	}
}

func TestStartRunsSetup(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	fsys["wl-scale/setup.sh"] = &fstest.MapFile{Data: []byte("kubectl create ns wl-scale\n")}
	r := &fakeRunner{}

	err := Start(context.Background(), fsys, r, Task{ID: "wl-scale", Dir: "wl-scale"})

	if err != nil {
		t.Fatal(err)
	}
	if r.taskID != "wl-scale" || r.script != "kubectl create ns wl-scale\n" {
		t.Errorf("ran %q for %q", r.script, r.taskID)
	}
}

func TestStartReportsSetupFailure(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	r := &fakeRunner{err: errors.New("namespace stuck")}

	err := Start(context.Background(), fsys, r, Task{ID: "wl-scale", Dir: "wl-scale"})

	if err == nil || !strings.Contains(err.Error(), "set up wl-scale: namespace stuck") {
		t.Errorf("got error %v", err)
	}
}

func taskMD(id, fields string) string {
	return "---\nid: " + id + "\ntitle: A task\n" + fields + "\n---\n"
}

func addTask(fsys fstest.MapFS, id string) {
	fsys[id+"/task.md"] = &fstest.MapFile{Data: []byte(taskMD(id, "domain: workloads\nweight: 1"))}
	for _, script := range Scripts {
		fsys[id+"/"+script] = &fstest.MapFile{Data: []byte("true\n")}
	}
	fsys[id+"/explain.md"] = &fstest.MapFile{Data: []byte("Why.\n")}
}

type fakeRunner struct {
	taskID, script string
	err            error
}

func (r *fakeRunner) Run(_ context.Context, taskID string, script []byte) (string, error) {
	r.taskID, r.script = taskID, string(script)
	return "", r.err
}
