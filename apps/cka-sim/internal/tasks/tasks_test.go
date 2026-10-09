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

	want := Task{ID: "wl-scale", Title: "A task", Domain: Workloads, Topics: []string{"deployments", "scaling"}, Weight: 4, Host: "node-1", Question: "## Task\n\nScale it."}
	if task.ID != want.ID || task.Title != want.Title || task.Domain != want.Domain ||
		!slices.Equal(task.Topics, want.Topics) || task.Weight != want.Weight || task.Host != want.Host || task.Question != want.Question {
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
		{"no title", "---\nid: x\nhost: n\ndomain: workloads\nweight: 1\n---\n", "title is required"},
		{"no host", "---\nid: x\ntitle: A task\ndomain: workloads\nweight: 1\n---\n", "host is required"},
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
	r := &fakeRunner{}

	err := Start(context.Background(), fsys, r, wlScale)

	if err != nil {
		t.Fatal(err)
	}
	want := []call{{"node-1", "wl-scale", "setup.sh"}}
	if !slices.Equal(r.calls, want) {
		t.Errorf("ran %+v, want %+v", r.calls, want)
	}
}

func TestStartReportsSetupFailure(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	r := &fakeRunner{err: errors.New("namespace stuck")}

	err := Start(context.Background(), fsys, r, wlScale)

	if err == nil || !strings.Contains(err.Error(), "set up wl-scale: namespace stuck") {
		t.Errorf("got error %v", err)
	}
}

func TestCheckGradesTheOutput(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	r := &fakeRunner{out: map[string]string{"check.sh": "noise\nPASS 2 scaled\nFAIL 3 ready\n"}}

	result, err := Check(context.Background(), fsys, r, wlScale)

	if err != nil {
		t.Fatal(err)
	}
	if want := []call{{"node-1", "wl-scale", "check.sh"}}; !slices.Equal(r.calls, want) {
		t.Errorf("ran %+v, want %+v", r.calls, want)
	}
	if result.Earned != 2 || result.Total != 5 || len(result.Checks) != 2 {
		t.Errorf("got %+v, want 2/5 from two checks", result)
	}
}

func TestCheckReportsScriptFailure(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	r := &fakeRunner{err: errors.New("kubectl: not found")}

	_, err := Check(context.Background(), fsys, r, wlScale)

	if err == nil || !strings.Contains(err.Error(), "check wl-scale: kubectl: not found") {
		t.Errorf("got error %v", err)
	}
}

func TestSolveRunsSolution(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	r := &fakeRunner{}

	err := Solve(context.Background(), fsys, r, wlScale)

	if err != nil {
		t.Fatal(err)
	}
	if want := []call{{"node-1", "wl-scale", "solution.sh"}}; !slices.Equal(r.calls, want) {
		t.Errorf("ran %+v, want %+v", r.calls, want)
	}
}

func TestSelftestPassesAFairTask(t *testing.T) {
	fsys := fstest.MapFS{}
	addTask(fsys, "wl-scale")
	r := &fakeRunner{
		out:       map[string]string{"check.sh": "FAIL 2 scaled\nFAIL 2 ready\n"},
		solvedOut: map[string]string{"check.sh": "PASS 2 scaled\nPASS 2 ready\n"},
	}

	err := Selftest(context.Background(), fsys, r, wlScale)

	if err != nil {
		t.Fatal(err)
	}
	var scripts []string
	for _, c := range r.calls {
		scripts = append(scripts, c.script)
	}
	if want := []string{"setup.sh", "check.sh", "solution.sh", "check.sh"}; !slices.Equal(scripts, want) {
		t.Errorf("ran %q, want %q", scripts, want)
	}
}

func TestSelftestRejects(t *testing.T) {
	tests := []struct {
		name          string
		before, after string
		wantErr       string
	}{
		{"points before the solution", "PASS 2 scaled\nFAIL 2 ready\n", "PASS 2 scaled\nPASS 2 ready\n", `earns 2/4 before the solution: "scaled"`},
		{"points missing after the solution", "FAIL 2 scaled\nFAIL 2 ready\n", "PASS 2 scaled\nFAIL 2 ready\n", `earns 2/4 after the solution: "ready"`},
		{"no checks", "all good\n", "all good\n", "check.sh prints no checks"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{}
			addTask(fsys, "wl-scale")
			r := &fakeRunner{out: map[string]string{"check.sh": tt.before}, solvedOut: map[string]string{"check.sh": tt.after}}

			err := Selftest(context.Background(), fsys, r, wlScale)

			if err == nil || !strings.Contains(err.Error(), "wl-scale: "+tt.wantErr) {
				t.Errorf("got error %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

var wlScale = Task{ID: "wl-scale", Dir: "wl-scale", Host: "node-1"}

func taskMD(id, fields string) string {
	return "---\nid: " + id + "\ntitle: A task\nhost: node-1\n" + fields + "\n---\n"
}

func addTask(fsys fstest.MapFS, id string) {
	fsys[id+"/task.md"] = &fstest.MapFile{Data: []byte(taskMD(id, "domain: workloads\nweight: 1"))}
	for _, script := range Scripts {
		fsys[id+"/"+script] = &fstest.MapFile{Data: []byte(script + "\n")}
	}
	fsys[id+"/explain.md"] = &fstest.MapFile{Data: []byte("Why.\n")}
}

type call struct{ host, taskID, script string }

// fakeRunner answers by script name (each fake script's body is its own name); after solution.sh
// has run, solvedOut wins over out.
type fakeRunner struct {
	calls          []call
	out, solvedOut map[string]string
	solved         bool
	err            error
}

func (r *fakeRunner) Run(_ context.Context, host, taskID string, script []byte) (string, error) {
	name := strings.TrimSpace(string(script))
	r.calls = append(r.calls, call{host, taskID, name})
	if name == "solution.sh" {
		r.solved = true
	}
	if out, ok := r.solvedOut[name]; ok && r.solved {
		return out, r.err
	}
	return r.out[name], r.err
}
