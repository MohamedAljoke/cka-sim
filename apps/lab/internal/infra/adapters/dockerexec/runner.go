package dockerexec

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
)

const (
	outputLimit = 64 << 10
	stopDelay   = 5 * time.Second
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, host, script string) (cluster.ScriptResult, error) {
	pidFile := "/tmp/cka-sim-run-" + rand.Text()
	// setsid makes the script lead its own process group, so a cancel can stop it and everything it started.
	//nolint:gosec // no shell runs the arguments, and the host is one of the cluster's nodes
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", "-e", "KUBECONFIG=/etc/kubernetes/admin.conf",
		host, "setsid", "-w", "bash", "-s")
	cmd.Stdin = io.MultiReader(
		strings.NewReader("echo $$ > "+pidFile+"; trap 'rm -f "+pidFile+"' EXIT\n"),
		strings.NewReader(script))

	// Killing the docker client alone would leave the script running inside the node.
	cmd.Cancel = func() error {
		stop := `kill -9 "-$(cat ` + pidFile + `)" 2>/dev/null; rm -f ` + pidFile
		//nolint:gosec // no shell runs the arguments, and the pid file is ours
		_ = exec.CommandContext(context.WithoutCancel(ctx), "docker", "exec", host, "sh", "-c", stop).Run()

		return cmd.Process.Kill()
	}
	cmd.WaitDelay = stopDelay

	output := &lastBytes{limit: outputLimit}
	cmd.Stdout = output
	cmd.Stderr = output

	err := cmd.Run()

	if ctx.Err() != nil {
		return cluster.ScriptResult{}, fmt.Errorf("the script on %s stopped: %w", host, ctx.Err())
	}

	var exited *exec.ExitError
	if err != nil && !errors.As(err, &exited) {
		return cluster.ScriptResult{}, fmt.Errorf("run a script on %s: %w", host, err)
	}

	return cluster.ScriptResult{Output: output.buf, ExitCode: cmd.ProcessState.ExitCode()}, nil
}

// The end of a script's output says why it failed, so a long one keeps its last bytes.
type lastBytes struct {
	buf   []byte
	limit int
}

func (b *lastBytes) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	if extra := len(b.buf) - b.limit; extra > 0 {
		b.buf = b.buf[extra:]
	}

	return len(p), nil
}
