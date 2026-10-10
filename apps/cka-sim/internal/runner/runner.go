package runner

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type Node struct {
	Lib []byte
	// DockerHost is the daemon the nodes run on; empty means docker's own default.
	DockerHost string
}

func (n Node) Run(ctx context.Context, host, taskID string, script []byte) (string, error) {
	pidFile := "/tmp/cka-sim-run-" + rand.Text()
	// setsid makes the script lead its own process group, so a cancel can stop it and everything it started.
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", "-e", "TASK_ID="+taskID, host, "setsid", "-w", "bash", "-s")
	cmd.Env = n.env()
	cmd.Stdin = io.MultiReader(
		bytes.NewReader([]byte("echo $$ > "+pidFile+"; trap 'rm -f "+pidFile+"' EXIT\n")),
		bytes.NewReader(n.Lib), bytes.NewReader([]byte("\n")), bytes.NewReader(script))
	// Killing the docker client alone would leave the script running inside the node.
	cmd.Cancel = func() error {
		stop := `kill -9 "-$(cat ` + pidFile + `)" 2>/dev/null; rm -f ` + pidFile
		kill := exec.Command("docker", "exec", host, "sh", "-c", stop)
		kill.Env = n.env()
		kill.Run()
		return cmd.Process.Kill()
	}
	cmd.WaitDelay = 5 * time.Second
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return string(out), fmt.Errorf("script for %s on %s stopped: %w", taskID, host, ctx.Err())
	}
	if err != nil {
		return string(out), fmt.Errorf("run script for %s on %s: %w\n%s", taskID, host, err, out)
	}
	return string(out), nil
}

func (n Node) env() []string {
	if n.DockerHost == "" {
		return nil
	}
	return append(os.Environ(), "DOCKER_HOST="+n.DockerHost)
}
