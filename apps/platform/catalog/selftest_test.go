package catalog

import (
	"cmp"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

// Opt-in: it rebuilds every task's namespace on the running lab and takes minutes.
var selftest = flag.Bool("selftest", false, "run every task against the running lab")

func TestSelftest(t *testing.T) {
	if !*selftest {
		t.Skip("needs -selftest and a running lab; run make selftest")
	}
	lab := labagent.Client{URL: cmp.Or(os.Getenv("CKA_SIM_LAB_URL"), "http://127.0.0.1:7171"), Token: os.Getenv("CKA_SIM_AGENT_TOKEN")}
	if err := lab.Healthy(context.Background()); err != nil {
		t.Skipf("no lab answers; run make up: %v", err)
	}
	all, err := tasks.Load(FS)
	if err != nil {
		t.Fatal(err)
	}

	for _, task := range all {
		t.Run(task.ID, func(t *testing.T) {
			// A failed selftest would otherwise leave a broken node behind for the next task, and every
			// task's leftover Pods together would fill the workers' Pod limit.
			t.Cleanup(func() {
				if err := tasks.Tidy(context.Background(), FS, lab, []tasks.Task{task}); err != nil {
					t.Error(err)
				}
			})
			if err := tasks.Selftest(context.Background(), FS, lab, task); err != nil {
				t.Error(err)
			}
		})
	}
}
