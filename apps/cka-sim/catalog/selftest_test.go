package catalog

import (
	"context"
	"flag"
	"io/fs"
	"os/exec"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/runner"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

// Opt-in: it rebuilds every task's namespace on the running cluster and takes minutes.
var selftest = flag.Bool("selftest", false, "run every task against the running cluster")

func TestSelftest(t *testing.T) {
	if !*selftest {
		t.Skip("needs -selftest and a running cluster; run make selftest")
	}
	all, err := tasks.Load(FS)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := fs.ReadFile(FS, "lib.sh")
	if err != nil {
		t.Fatal(err)
	}
	node := runner.Node{Lib: lib}

	for _, task := range all {
		t.Run(task.ID, func(t *testing.T) {
			if exec.Command("docker", "inspect", task.Host).Run() != nil {
				t.Skipf("%s is not running; run make up", task.Host)
			}
			// A failed selftest would otherwise leave a broken node behind for the next task.
			t.Cleanup(func() {
				if err := tasks.Heal(context.Background(), FS, node, []tasks.Task{task}); err != nil {
					t.Error(err)
				}
			})
			if err := tasks.Selftest(context.Background(), FS, node, task); err != nil {
				t.Error(err)
			}
		})
	}
}
