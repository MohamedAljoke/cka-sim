package catalog

import (
	"context"
	"io/fs"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/files"
)

func TestEveryTaskLoads(t *testing.T) {
	repo, err := files.NewTaskRepository(FS)
	if err != nil {
		t.Fatal(err)
	}

	all, _ := repo.List(context.Background())

	if len(all) == 0 {
		t.Error("the catalog has no tasks")
	}
}

func TestEveryResetIsLoaded(t *testing.T) {
	resets, err := fs.Glob(FS, "*/reset.sh")
	if err != nil {
		t.Fatal(err)
	}
	repo, err := files.NewTaskRepository(FS)
	if err != nil {
		t.Fatal(err)
	}

	all, _ := repo.List(context.Background())

	loaded := 0
	for _, task := range all {
		if task.Reset != "" {
			loaded++
		}
	}
	if len(resets) == 0 || loaded != len(resets) {
		t.Errorf("%d reset.sh files, %d tasks with a reset", len(resets), loaded)
	}
}
