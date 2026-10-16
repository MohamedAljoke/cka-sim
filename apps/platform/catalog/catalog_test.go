package catalog

import (
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

func TestEveryTaskLoads(t *testing.T) {
	all, err := tasks.Load(FS)

	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Error("the catalog has no tasks")
	}
}
