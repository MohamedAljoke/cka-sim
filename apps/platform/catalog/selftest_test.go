package catalog

import (
	"cmp"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/files"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/labagent"
)

// Opt-in: it sets up and solves every task on the running lab, which takes a long time.
var selftest = flag.Bool("selftest", false, "set up, check, solve and check every task on the running lab")

// The selftest runs on one person's own lab, as the local student.
const owner = "local"

func TestSelftest(t *testing.T) {
	if !*selftest {
		t.Skip("needs -selftest and a running lab; run make selftest")
	}

	container := labagent.Fixed(cmp.Or(os.Getenv("CKA_SIM_LAB_URL"), "http://127.0.0.1:7171"))
	lab := labagent.NewFleet(func(string) labagent.Machine { return container }, os.Getenv("CKA_SIM_AGENT_TOKEN"))

	if got, err := lab.Get(context.Background(), owner); err != nil || got.State != labs.Ready {
		t.Skipf("the lab is not ready (%+v, %v); start it and set CKA_SIM_AGENT_TOKEN", got, err)
	}

	repo, err := files.NewTaskRepository(FS)
	if err != nil {
		t.Fatal(err)
	}

	uc := usecases.New(repo, lab)
	all, _ := repo.List(context.Background())

	for _, task := range all {
		t.Run(task.ID, func(t *testing.T) {
			// A failed task would otherwise leave a broken node for the next one, and the leftover Pods of
			// every task together would fill the workers' Pod limit.
			t.Cleanup(func() {
				if err := uc.Tidy.Execute(context.Background(), owner, task.ID); err != nil {
					t.Error(err)
				}
			})

			if err := uc.Selftest.Execute(context.Background(), owner, task.ID); err != nil {
				t.Error(err)
			}
		})
	}
}
