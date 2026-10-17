package kind

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	kindcluster "sigs.k8s.io/kind/pkg/cluster"
)

//go:embed kind.yaml
var config string

//go:embed kind-cgroupv1-patch.yaml
var cgroupV1Patch string

type Provisioner struct {
	name       string
	kubeconfig string
}

func NewProvisioner(name string) (*Provisioner, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("find a folder for the kubeconfig: %w", err)
	}

	dir := filepath.Join(cache, "cka-sim")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	return &Provisioner{name: name, kubeconfig: filepath.Join(dir, name+".kubeconfig")}, nil
}

func (p *Provisioner) Create(ctx context.Context, progress func(step string)) error {
	if err := p.Delete(ctx); err != nil {
		return fmt.Errorf("clear what an earlier try left: %w", err)
	}

	cgroupVersion, err := dockerCgroupVersion(ctx)
	if err != nil {
		return err
	}

	err = newProvider(progress).Create(p.name,
		kindcluster.CreateWithRawConfig([]byte(kindConfig(cgroupVersion))),
		kindcluster.CreateWithKubeconfigPath(p.kubeconfig),
		kindcluster.CreateWithWaitForReady(5*time.Minute),
		kindcluster.CreateWithDisplayUsage(false),
		kindcluster.CreateWithDisplaySalutation(false),
	)
	if err != nil {
		return err
	}

	progress("Waiting for every node to be Ready")

	return p.waitForNodes(ctx)
}

func (p *Provisioner) Delete(context.Context) error {
	return newProvider(func(string) {}).Delete(p.name, p.kubeconfig)
}

func (p *Provisioner) Nodes(context.Context) ([]string, error) {
	nodes, err := newProvider(func(string) {}).ListNodes(p.name)
	if err != nil {
		return nil, fmt.Errorf("list the nodes of %s: %w", p.name, err)
	}

	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.String())
	}

	return names, nil
}

func newProvider(progress func(string)) *kindcluster.Provider {
	return kindcluster.NewProvider(kindcluster.ProviderWithDocker(), kindcluster.ProviderWithLogger(stepLogger{progress}))
}

// kind only waits for the control plane; the workers join after it.
func (p *Provisioner) waitForNodes(ctx context.Context) error {
	//nolint:gosec // no shell runs the arguments, and the name comes from the agent's own settings
	out, err := exec.CommandContext(ctx, "docker", "exec", p.name+"-control-plane",
		"kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"wait", "--for=condition=Ready", "nodes", "--all", "--timeout=5m").CombinedOutput()
	if err != nil {
		return fmt.Errorf("wait for the nodes: %w: %s", err, out)
	}

	return nil
}

func kindConfig(dockerCgroupVersion string) string {
	if dockerCgroupVersion == "1" {
		return config + cgroupV1Patch
	}

	return config
}

func dockerCgroupVersion(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "docker", "info", "--format", "{{.CgroupVersion}}").Output()
	if err != nil {
		return "", fmt.Errorf("read docker's cgroup version: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
