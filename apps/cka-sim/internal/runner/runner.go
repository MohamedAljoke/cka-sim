package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type Node struct {
	Lib []byte
}

func (n Node) Run(ctx context.Context, host, taskID string, script []byte) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", "-e", "TASK_ID="+taskID, host, "bash", "-s")
	cmd.Stdin = io.MultiReader(bytes.NewReader(n.Lib), bytes.NewReader([]byte("\n")), bytes.NewReader(script))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("run script for %s on %s: %w\n%s", taskID, host, err, out)
	}
	return string(out), nil
}
