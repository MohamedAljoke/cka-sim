package tasks

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"testing/fstest"

	ckasim "github.com/MohamedAljoke/cka-sim"
)

func TestParse(t *testing.T) {
	task, err := Parse("---\nid: tr-x\ntitle: Broken thing\ndomain: troubleshooting\nweight: 7\ncluster: cka1\nhost: cka1-worker\norder: last\n---\n## Task\n\nFix it.\n")
	if err != nil {
		t.Fatal(err)
	}
	want := Task{ID: "tr-x", Title: "Broken thing", Domain: Troubleshooting, Weight: 7, Cluster: "cka1", Host: "cka1-worker", Last: true, Body: "## Task\n\nFix it."}
	if task != want {
		t.Fatalf("got %+v\nwant %+v", task, want)
	}
}

func TestParseRejectsIncompleteTasks(t *testing.T) {
	for name, text := range map[string]string{
		"no frontmatter": "## Task",
		"unterminated":   "---\nid: x\n",
		"no host":        "---\nid: x\ndomain: storage\nweight: 1\ncluster: c\n---\n",
		"bad domain":     "---\nid: x\ndomain: magic\nweight: 1\ncluster: c\nhost: h\n---\n",
		"bad weight":     "---\nid: x\ndomain: storage\nweight: lots\ncluster: c\nhost: h\n---\n",
	} {
		if _, err := Parse(text); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func pool(perDomain int) []Task {
	var all []Task
	for d := range DomainWeights {
		for i := range perDomain {
			all = append(all, Task{ID: fmt.Sprintf("%s-%d", d, i), Domain: d})
		}
	}
	return all
}

func TestDrawFollowsDomainWeights(t *testing.T) {
	exam := Draw(pool(10), 20, rand.New(rand.NewPCG(1, 2)))
	count := map[Domain]int{}
	seen := map[string]bool{}
	for _, task := range exam {
		count[task.Domain]++
		if seen[task.ID] {
			t.Fatalf("task %s drawn twice", task.ID)
		}
		seen[task.ID] = true
	}
	want := map[Domain]int{Troubleshooting: 6, Architecture: 5, Networking: 4, Workloads: 3, Storage: 2}
	for d, n := range want {
		if count[d] != n {
			t.Errorf("%s: got %d tasks, want %d (all: %v)", d, count[d], n, count)
		}
	}
}

func TestDrawNeverExceedsThePool(t *testing.T) {
	small := pool(1)
	for _, n := range []int{1, 3, 5, 16} {
		exam := Draw(small, n, rand.New(rand.NewPCG(3, 4)))
		if want := min(n, len(small)); len(exam) != want {
			t.Errorf("n=%d: got %d tasks, want %d", n, len(exam), want)
		}
	}
}

// Every task shipped in the binary must parse, with a known host and a full script set.
func TestShippedTasksAreComplete(t *testing.T) {
	all, err := Load(ckasim.Assets)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("no tasks found")
	}
	for _, task := range all {
		for _, f := range []string{"setup.sh", "check.sh", "solution.sh", "explain.md"} {
			if _, err := ckasim.Assets.Open(task.Dir + "/" + f); err != nil {
				t.Errorf("%s: missing %s", task.ID, f)
			}
		}
		if task.Explain == "" {
			t.Errorf("%s: empty explanation", task.ID)
		}
	}
}

func TestLoadReadsExplain(t *testing.T) {
	fsys := fstest.MapFS{
		"tasks/storage/st-a/task.md":    {Data: []byte("---\nid: st-a\ndomain: storage\nweight: 3\ncluster: c\nhost: h\n---\nbody")},
		"tasks/storage/st-a/explain.md": {Data: []byte("why")},
	}
	all, err := Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].Explain != "why" || all[0].Dir != "tasks/storage/st-a" {
		t.Fatalf("unexpected: %+v", all)
	}
}
