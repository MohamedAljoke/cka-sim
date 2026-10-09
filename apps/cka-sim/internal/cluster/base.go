package cluster

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"strings"
)

//go:embed base.Dockerfile
var baseDockerfile string

// Prepare gives the cluster its exam layout: a base host to work from, and ssh access from it
// to every node as candidate. An existing base is left alone, so open terminals stay open.
func (c *Cluster) Prepare(ctx context.Context) error {
	if baseExists() {
		return nil
	}
	if err := buildImage(ctx, BaseImage, baseDockerfile); err != nil {
		return err
	}
	key, err := startBase(ctx)
	if err != nil {
		return err
	}
	return provisionNodes(ctx, key)
}

// --network kind makes the node names resolve; --init reaps what ended shells leave behind.
func startBase(ctx context.Context) (publicKey string, err error) {
	if _, err := docker(ctx, "", "rm", "-f", Base); err != nil {
		return "", err
	}
	if _, err := docker(ctx, "", "run", "-d", "--init", "--name", Base, "--hostname", "base",
		"--network", "kind", "--label", "io.x-k8s.kind.cluster="+Name, BaseImage); err != nil {
		return "", err
	}
	return docker(ctx, "", "exec", "-u", Candidate, Base, "bash", "-c",
		`install -d -m 700 ~/.ssh && ssh-keygen -q -t ed25519 -N "" -C candidate@base -f ~/.ssh/id_ed25519 && cat ~/.ssh/id_ed25519.pub`)
}

func provisionNodes(ctx context.Context, publicKey string) error {
	kubeconfig, err := docker(ctx, "", "exec", ControlPlaneNode, "cat", "/etc/kubernetes/admin.conf")
	if err != nil {
		return err
	}
	home := "/home/" + Candidate
	for _, node := range Nodes {
		if _, err := docker(ctx, "", "exec", node, "id", Candidate); err != nil {
			return fmt.Errorf("node %s has no %s user, so it runs an older node image; run cka-sim down, then cka-sim up", node, Candidate)
		}
		if _, err := docker(ctx, "", "exec", node, "install", "-d", "-o", Candidate, "-g", Candidate, "-m", "700", home+"/.kube", home+"/.ssh"); err != nil {
			return err
		}
		if _, err := docker(ctx, kubeconfig, "exec", "-i", node, "install", "-o", Candidate, "-g", Candidate, "-m", "600", "/dev/stdin", home+"/.kube/config"); err != nil {
			return err
		}
		if _, err := docker(ctx, publicKey, "exec", "-i", node, "install", "-o", Candidate, "-g", Candidate, "-m", "600", "/dev/stdin", home+"/.ssh/authorized_keys"); err != nil {
			return err
		}
		// Task scripts run as root on the task's host, and only the control plane has a kubeconfig.
		if _, err := docker(ctx, kubeconfig, "exec", "-i", node, "install", "-D", "-m", "600", "/dev/stdin", "/root/.kube/config"); err != nil {
			return err
		}
	}
	return nil
}

func removeBase(ctx context.Context) error {
	_, err := docker(ctx, "", "rm", "-f", Base)
	return err
}

func baseExists() bool {
	return exec.Command("docker", "inspect", Base).Run() == nil
}

func docker(ctx context.Context, stdin string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return string(out), nil
}
