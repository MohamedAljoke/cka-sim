package files

import (
	"context"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestParseReadsFrontmatterAndQuestion(t *testing.T) {
	task, err := parseTaskFile(taskMD("wl-scale", "domain: workloads\ntopics: [deployments, scaling]\nweight: 4\norder: last") + "\n## Task\n\nScale it.\n")
	if err != nil {
		t.Fatal(err)
	}

	want := tasks.Task{ID: "wl-scale", Title: "A task", Domain: tasks.Workloads, Topics: []string{"deployments", "scaling"}, Weight: 4, Host: "node-1", Order: tasks.Last, Question: "## Task\n\nScale it."}
	if task.ID != want.ID || task.Title != want.Title || task.Domain != want.Domain || !slices.Equal(task.Topics, want.Topics) ||
		task.Weight != want.Weight || task.Host != want.Host || task.Order != want.Order || task.Question != want.Question {
		t.Errorf("got %+v\nwant %+v", task, want)
	}
}

func TestParseRejects(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr string
	}{
		{"no taskFrontmatter", "## Task\n", "missing taskFrontmatter"},
		{"unclosed taskFrontmatter", "---\nid: x\n", "no closing ---"},
		{"unknown key", taskMD("x", "domain: workloads\nweight: 1\ncluster: a"), "unknown field"},
		{"wrong type", taskMD("x", "domain: workloads\nweight: many"), "taskFrontmatter:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTaskFile(tt.text)

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got error %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestListGivesEveryTaskFile(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	addTask(fsys, "net-svc")
	fsys["notes.md"] = &fstest.MapFile{Data: []byte("not a task")}
	repo, err := NewTaskRepository(fsys)
	if err != nil {
		t.Fatal(err)
	}

	all, err := repo.List(context.Background())

	if err != nil || len(all) != 2 {
		t.Errorf("List = %d tasks, %v; want 2", len(all), err)
	}
}

func TestListIsACopy(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	repo, err := NewTaskRepository(fsys)
	if err != nil {
		t.Fatal(err)
	}

	first, _ := repo.List(context.Background())
	first[0].Title = "changed"
	second, _ := repo.List(context.Background())

	if second[0].Title == "changed" {
		t.Error("changing what List returned changed the repository")
	}
}

func TestNewTaskFilesRejects(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr string
	}{
		{"bad taskFrontmatter names the file", "## Task\n", "wl-scale/task.md: missing taskFrontmatter"},
		{"id differs from folder", taskMD("wl-other", "domain: workloads\nweight: 1"), `id "wl-other" must match its folder "wl-scale"`},
		{"a broken task rule names the file", taskMD("wl-scale", "domain: security\nweight: 1"), `wl-scale/task.md: unknown domain "security"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{"wl-scale/task.md": &fstest.MapFile{Data: []byte(tt.text)}}

			_, err := NewTaskRepository(fsys)

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
