package catalog

import (
	"io/fs"
	"path"
	"strings"
	"testing"

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
		for _, script := range tasks.Scripts {
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
