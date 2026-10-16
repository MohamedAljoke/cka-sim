package catalog

import (
	"context"
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
