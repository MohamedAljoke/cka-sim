package tasks

import (
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

	want := Task{ID: "wl-scale", Title: "A task", Domain: Workloads, Topics: []string{"deployments", "scaling"}, Weight: 4, Host: "node-1", Question: "## Task\n\nScale it."}
	if task.ID != want.ID || task.Title != want.Title || task.Domain != want.Domain ||
		!slices.Equal(task.Topics, want.Topics) || task.Weight != want.Weight || task.Host != want.Host || task.Question != want.Question {
		t.Errorf("got %+v\nwant %+v", task, want)
	}
}

func TestParseOrderLast(t *testing.T) {
	task, err := Parse(taskMD("ar-etcd", "domain: architecture\nweight: 8\norder: last"))
	if err != nil {
		t.Fatal(err)
	}

	if task.Order != Last {
		t.Errorf("order %q, want last", task.Order)
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
		{"no title", "---\nid: x\nhost: n\ndomain: workloads\nweight: 1\n---\n", "title is required"},
		{"no host", "---\nid: x\ntitle: A task\ndomain: workloads\nweight: 1\n---\n", "host is required"},
		{"unknown domain", taskMD("x", "domain: security\nweight: 1"), `unknown domain "security"`},
		{"no weight", taskMD("x", "domain: workloads"), "weight must be more than 0"},
		{"capital topic", taskMD("x", "domain: workloads\nweight: 1\ntopics: [Deployments]"), "lowercase-kebab"},
		{"spaced topic", taskMD("x", "domain: workloads\nweight: 1\ntopics: [rolling update]"), "lowercase-kebab"},
		{"unknown order", taskMD("x", "domain: workloads\nweight: 1\norder: first"), `order must be empty or "last"`},
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
}

func TestLoadRejects(t *testing.T) {
	tests := []struct {
		name    string
		breakIt func(fstest.MapFS)
		wantErr string
	}{
		{"bad frontmatter names the file", func(f fstest.MapFS) {
			f["wl-scale/task.md"] = &fstest.MapFile{Data: []byte(taskMD("wl-scale", "domain: security\nweight: 1"))}
		}, `wl-scale/task.md: unknown domain "security"`},
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

func taskMD(id, fields string) string {
	return "---\nid: " + id + "\ntitle: A task\nhost: node-1\n" + fields + "\n---\n"
}

func addTask(fsys fstest.MapFS, id string) {
	fsys[id+"/task.md"] = &fstest.MapFile{Data: []byte(taskMD(id, "domain: workloads\nweight: 1"))}
}
