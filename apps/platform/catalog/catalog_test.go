package catalog

import (
	"io/fs"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/eval"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
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
		if !slices.Contains(labagent.Nodes, task.Host) {
			t.Errorf("%s: host %q is not a node, want one of %q", task.ID, task.Host, labagent.Nodes)
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

// The node image already has registry.k8s.io's images and the lab's registry serves its own; the rest must be
// in images.txt, which make up and the GKE pool hand to the lab to preload.
var imageRef = regexp.MustCompile(`(?:--image=|image: )([\w./:-]+)`)

func TestEveryImageIsPreloaded(t *testing.T) {
	list, err := os.ReadFile("images.txt")
	if err != nil {
		t.Fatal(err)
	}
	preloaded := strings.Fields(string(list))
	scripts, err := fs.Glob(FS, "*/*.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, script := range scripts {
		body, _ := fs.ReadFile(FS, script)
		for _, m := range imageRef.FindAllStringSubmatch(string(body), -1) {
			if image := m[1]; !strings.HasPrefix(image, "registry.k8s.io/") && !strings.HasPrefix(image, "cka-sim-registry:5000/") && !slices.Contains(preloaded, image) {
				t.Errorf("%s runs %s, which images.txt doesn't preload", script, image)
			}
		}
	}
}

func TestEveryEvalSetLoads(t *testing.T) {
	all, err := tasks.Load(FS)
	if err != nil {
		t.Fatal(err)
	}
	set, err := eval.Load(FS, all)
	if err != nil {
		t.Fatal(err)
	}
	if len(set) == 0 {
		t.Error("no task has an evals.yaml")
	}
	sets, _ := fs.Glob(FS, "*/evals.yaml")
	for _, file := range sets {
		if _, ok := tasks.Find(all, path.Dir(file)); !ok {
			t.Errorf("%s belongs to no task", file)
		}
	}
}

func TestLibIsShipped(t *testing.T) {
	if _, err := fs.Stat(FS, "lib.sh"); err != nil {
		t.Fatal(err)
	}
}
