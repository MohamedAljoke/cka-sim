package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type Node struct {
	Name string
	Lib  []byte
}

func (n Node) Run(ctx context.Context, taskID string, script []byte) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", "-e", "TASK_ID="+taskID, n.Name, "bash", "-s")
	cmd.Stdin = io.MultiReader(bytes.NewReader(n.Lib), bytes.NewReader([]byte("\n")), bytes.NewReader(script))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("run script for %s in %s: %w\n%s", taskID, n.Name, err, out)
	}
	return string(out), nil
}
