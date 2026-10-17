package files

import (
	"context"
	"errors"
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

	want := tasks.Task{ID: "wl-scale", Title: "A task", Domain: tasks.Workloads, Topics: []string{"deployments", "scaling"}, Weight: 4, Host: "node-1", Order: tasks.Last, Question: "Do the wl-scale task.\n\n## Task\n\nScale it."}
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
		{"no frontmatter", "## Task\n", "missing frontmatter"},
		{"unclosed frontmatter", "---\nid: x\n", "no closing ---"},
		{"unknown key", taskMD("x", "domain: workloads\nweight: 1\ncluster: a"), "unknown field"},
		{"wrong type", taskMD("x", "domain: workloads\nweight: many"), "frontmatter:"},
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
	fsys := catalogFS()
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
	fsys := catalogFS()
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

func TestGetGivesTheTaskWithItsQuestionAndSolution(t *testing.T) {
	fsys := catalogFS()
	addTask(fsys, "wl-scale")
	repo, err := NewTaskRepository(fsys)
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.Get(context.Background(), "wl-scale")

	if err != nil || got.ID != "wl-scale" || got.Question != "Do the wl-scale task." || got.Solution != "Solve wl-scale." {
		t.Errorf("Get = %+v, %v", got, err)
	}
}

func TestGetIsNotFoundForAnUnknownID(t *testing.T) {
	repo, err := NewTaskRepository(catalogFS())
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.Get(context.Background(), "wl-scale")

	if !errors.Is(err, tasks.ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestNewTaskRepositoryRejects(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr string
	}{
		{"bad frontmatter names the file", "## Task\n", "wl-scale/task.md: missing frontmatter"},
		{"id differs from folder", taskMD("wl-other", "domain: workloads\nweight: 1"), `id "wl-other" must match its folder "wl-scale"`},
		{"a broken task rule names the file", taskMD("wl-scale", "domain: security\nweight: 1"), `wl-scale/task.md: unknown domain "security"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := catalogFS()
			addTask(fsys, "wl-scale")
			fsys["wl-scale/task.md"] = &fstest.MapFile{Data: []byte(tt.text)}

			_, err := NewTaskRepository(fsys)

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got error %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestNewTaskRepositoryNeedsEachTasksFiles(t *testing.T) {
	for _, file := range []string{"explain.md", "setup.sh", "check.sh"} {
		fsys := catalogFS()
		addTask(fsys, "wl-scale")
		delete(fsys, "wl-scale/"+file)

		_, err := NewTaskRepository(fsys)

		if err == nil || err.Error() != "wl-scale/task.md: missing "+file {
			t.Errorf("without %s: got error %v", file, err)
		}
	}
}

func TestNewTaskRepositoryNeedsTheLib(t *testing.T) {
	fsys := catalogFS()
	addTask(fsys, "wl-scale")
	delete(fsys, "lib.sh")

	if _, err := NewTaskRepository(fsys); err == nil || err.Error() != "missing lib.sh" {
		t.Errorf("got error %v", err)
	}
}

func TestScriptsRunAfterTheLibWithTheirTaskID(t *testing.T) {
	fsys := catalogFS()
	addTask(fsys, "wl-scale")
	repo, err := NewTaskRepository(fsys)
	if err != nil {
		t.Fatal(err)
	}

	got, _ := repo.Get(context.Background(), "wl-scale")

	if want := "export TASK_ID='wl-scale'\nfresh_ns() { :; }\n\nfresh_ns wl-scale\n"; got.Setup != want {
		t.Errorf("setup %q, want %q", got.Setup, want)
	}
	if want := "export TASK_ID='wl-scale'\nfresh_ns() { :; }\n\ncheck 1 \"wl-scale done\" true\n"; got.Check != want {
		t.Errorf("check %q, want %q", got.Check, want)
	}
}

func TestEmptyScriptsAreRejected(t *testing.T) {
	for file, wantErr := range map[string]string{"setup.sh": "setup is required", "check.sh": "check is required"} {
		fsys := catalogFS()
		addTask(fsys, "wl-scale")
		fsys["wl-scale/"+file] = &fstest.MapFile{Data: []byte("\n")}

		if _, err := NewTaskRepository(fsys); err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Errorf("an empty %s: got error %v", file, err)
		}
	}
}

func taskMD(id, fields string) string {
	return "---\nid: " + id + "\ntitle: A task\nhost: node-1\n" + fields + "\n---\nDo the " + id + " task.\n"
}

func catalogFS() fstest.MapFS {
	return fstest.MapFS{"lib.sh": &fstest.MapFile{Data: []byte("fresh_ns() { :; }\n")}}
}

func addTask(fsys fstest.MapFS, id string) {
	fsys[id+"/task.md"] = &fstest.MapFile{Data: []byte(taskMD(id, "domain: workloads\nweight: 1"))}
	fsys[id+"/explain.md"] = &fstest.MapFile{Data: []byte("Solve " + id + ".\n")}
	fsys[id+"/setup.sh"] = &fstest.MapFile{Data: []byte("fresh_ns " + id + "\n")}
	fsys[id+"/check.sh"] = &fstest.MapFile{Data: []byte("check 1 \"" + id + " done\" true\n")}
}
