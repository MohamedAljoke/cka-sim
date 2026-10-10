package catalog

import (
	"io/fs"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

func TestEveryShippedTaskLoads(t *testing.T) {
	all, err := tasks.Load(FS)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("the catalog has no tasks")
	}
	for _, task := range all {
		if !slices.Contains(cluster.Nodes, task.Host) {
			t.Errorf("%s: host %q is not a node, want one of %q", task.ID, task.Host, cluster.Nodes)
		}
		scripts := tasks.Scripts
		if task.Reset {
			scripts = append(slices.Clone(scripts), tasks.ResetScript)
		}
		for _, script := range scripts {
			body, _ := fs.ReadFile(FS, path.Join(task.Dir, script))
			if strings.TrimSpace(string(body)) == "" {
				t.Errorf("%s/%s is empty", task.ID, script)
			}
		}
	}
}

func TestLibIsShipped(t *testing.T) {
	if _, err := fs.Stat(FS, "lib.sh"); err != nil {
		t.Fatal(err)
	}
}
