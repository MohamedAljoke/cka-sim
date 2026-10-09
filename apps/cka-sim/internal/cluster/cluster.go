package cluster

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	kindcluster "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

const Name = "cka-sim"

const ControlPlaneNode = Name + "-control-plane"

// The tag must match the kindest/node version in node.Dockerfile.
const NodeImage = "cka-sim/node:v1.37.0"

//go:embed kind.yaml
var baseKindConfig string

//go:embed kind-cgroupv1-patch.yaml
var allowCgroupV1Patch string

//go:embed node.Dockerfile
var nodeDockerfile string

func kindConfig(dockerCgroupVersion string) string {
	if dockerCgroupVersion == "1" {
		return baseKindConfig + allowCgroupV1Patch
	}
	return baseKindConfig
}

type Cluster struct {
	provider       *kindcluster.Provider
	KubeconfigPath string
}

func New() (*Cluster, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("find a folder for the kubeconfig: %w", err)
	}
	return &Cluster{
		provider: kindcluster.NewProvider(
			kindcluster.ProviderWithDocker(),
			kindcluster.ProviderWithLogger(kindcmd.NewLogger()),
		),
		KubeconfigPath: filepath.Join(configDir, "cka-sim", "kubeconfig"),
	}, nil
}

func (c *Cluster) Exists() (bool, error) {
	names, err := c.provider.List()
	if err != nil {
		return false, err
	}
	return slices.Contains(names, Name), nil
}

func RequireUp() error {
	c, err := New()
	if err != nil {
		return err
	}
	exists, err := c.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("cluster %s is not up; run cka-sim up first", Name)
	}
	return nil
}

func (c *Cluster) Create(ctx context.Context) error {
	cgroupVersion, err := dockerCgroupVersion(ctx)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.KubeconfigPath), 0o700); err != nil {
		return err
	}
	if err := buildNodeImage(ctx); err != nil {
		return err
	}
	return c.provider.Create(Name,
		kindcluster.CreateWithNodeImage(NodeImage),
		kindcluster.CreateWithRawConfig([]byte(kindConfig(cgroupVersion))),
		kindcluster.CreateWithKubeconfigPath(c.KubeconfigPath),
		kindcluster.CreateWithWaitForReady(5*time.Minute),
		kindcluster.CreateWithDisplayUsage(false),
		kindcluster.CreateWithDisplaySalutation(false),
	)
}

func dockerCgroupVersion(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "docker", "info", "--format", "{{.CgroupVersion}}").Output()
	if err != nil {
		return "", fmt.Errorf("read docker's cgroup version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func buildNodeImage(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "build", "--quiet", "--tag", NodeImage, "-")
	cmd.Stdin = strings.NewReader(nodeDockerfile)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build node image %s: %w\n%s", NodeImage, err, out)
	}
	return nil
}

func (c *Cluster) Delete() error {
	return c.provider.Delete(Name, c.KubeconfigPath)
}
