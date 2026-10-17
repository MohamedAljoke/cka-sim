package kind

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"strings"
)

//go:embed base.Dockerfile
var baseDockerfile string

const (
	baseImage = "cka-sim/base:latest"
	candidate = "candidate"
)

func (p *Provisioner) startBase(ctx context.Context) error {
	if err := buildImage(ctx, baseImage, baseDockerfile); err != nil {
		return err
	}

	// --network kind makes the node names resolve; --init reaps what ended shells leave behind.
	if _, err := docker(ctx, "", "run", "-d", "--init", "--name", p.Base(), "--hostname", "base",
		"--network", "kind", baseImage); err != nil {
		return err
	}

	key, err := docker(ctx, "", "exec", "-u", candidate, p.Base(), "bash", "-c",
		`install -d -m 700 ~/.ssh && ssh-keygen -q -t ed25519 -N "" -C candidate@base -f ~/.ssh/id_ed25519 `+
			`&& cat ~/.ssh/id_ed25519.pub`)
	if err != nil {
		return err
	}

	return p.trustBase(ctx, key)
}

// Task scripts run as root on the task's host, and only the control plane comes with a kubeconfig.
func (p *Provisioner) trustBase(ctx context.Context, publicKey string) error {
	kubeconfig, err := docker(ctx, "", "exec", p.name+"-control-plane", "cat", "/etc/kubernetes/admin.conf")
	if err != nil {
		return err
	}

	nodes, err := p.Nodes(ctx)
	if err != nil {
		return err
	}

	home := "/home/" + candidate
	files := []struct{ owner, path, content string }{
		{candidate, home + "/.kube/config", kubeconfig},
		{candidate, home + "/.ssh/authorized_keys", publicKey},
		{"root", "/root/.kube/config", kubeconfig},
	}

	for _, node := range nodes {
		if _, err := docker(ctx, "", "exec", node,
			"install", "-d", "-o", candidate, "-g", candidate, "-m", "700", home+"/.kube", home+"/.ssh"); err != nil {
			return err
		}

		for _, f := range files {
			if _, err := docker(ctx, f.content, "exec", "-i", node,
				"install", "-D", "-o", f.owner, "-g", f.owner, "-m", "600", "/dev/stdin", f.path); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildImage(ctx context.Context, tag, dockerfile string) error {
	if _, err := docker(ctx, dockerfile, "build", "--quiet", "--tag", tag, "-"); err != nil {
		return fmt.Errorf("build the image %s: %w", tag, err)
	}

	return nil
}

func docker(ctx context.Context, stdin string, args ...string) (string, error) {
	//nolint:gosec // no shell runs the arguments, and they come from the agent's own settings
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = strings.NewReader(stdin)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker %s: %w: %s", strings.Join(args, " "), err, out)
	}

	return string(out), nil
}
