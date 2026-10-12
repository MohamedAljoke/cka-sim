package cluster

import (
	"context"
	_ "embed"
	"errors"
	"time"
)

const (
	Registry = "cka-sim-registry"
	Charts   = "cka-sim-charts"

	RegistryUser     = "candidate"
	RegistryPassword = "cka-registry"

	registryImage      = "cka-sim/registry:latest"
	chartsImage        = "cka-sim/charts:latest"
	helmImage          = "alpine/helm:3.22.0"
	metricsServerImage = "registry.k8s.io/metrics-server/metrics-server:v0.9.0"
	metricsServerPath  = "/etc/cka-sim/addons/metrics-server.yaml"
)

//go:embed metrics-server.yaml
var metricsServerManifest string

//go:embed registry.Dockerfile
var registryDockerfile string

//go:embed charts.Dockerfile
var chartsDockerfile string

var registryImages = map[string]string{
	"nginx:1.26-alpine": "team/web:1.0",
	"nginx:1.27-alpine": "team/web:1.1",
}

const registryHosts = `server = "http://` + Registry + `:5000"

[host."http://` + Registry + `:5000"]
  capabilities = ["pull", "resolve"]
`

func (c *Cluster) InstallAddons(ctx context.Context) error {
	return errors.Join(
		c.installMetricsServer(ctx),
		installHelm(ctx),
		startRegistry(ctx),
		startAddon(ctx, Charts, chartsImage, chartsDockerfile),
	)
}

func (c *Cluster) installMetricsServer(ctx context.Context) error {
	if err := c.LoadImages(ctx, []string{metricsServerImage}); err != nil {
		return err
	}
	if _, err := docker(ctx, metricsServerManifest, "exec", "-i", ControlPlaneNode, "install", "-D", "-m", "644", "/dev/stdin", metricsServerPath); err != nil {
		return err
	}
	_, err := docker(ctx, "", "exec", ControlPlaneNode, "kubectl", "--kubeconfig", "/etc/kubernetes/admin.conf", "apply", "-f", metricsServerPath)
	return err
}

func installHelm(ctx context.Context) error {
	if err := pullMissing(ctx, []string{helmImage}); err != nil {
		return err
	}
	helm, err := docker(ctx, "", "run", "--rm", "--entrypoint", "cat", helmImage, "/usr/bin/helm")
	if err != nil {
		return err
	}
	return eachAtOnce(Nodes, func(node string) error {
		if _, err := docker(ctx, helm, "exec", "-i", node, "install", "-m", "755", "/dev/stdin", "/usr/local/bin/helm"); err != nil {
			return err
		}
		_, err := docker(ctx, "", "exec", node, "sh", "-c", "helm completion bash > /etc/bash_completion.d/helm")
		return err
	})
}

func startRegistry(ctx context.Context) error {
	if err := startAddon(ctx, Registry, registryImage, registryDockerfile, "-p", "127.0.0.1:5000:5000"); err != nil {
		return err
	}
	if err := eachAtOnce(Nodes, func(node string) error {
		_, err := docker(ctx, registryHosts, "exec", "-i", node, "install", "-D", "-m", "644", "/dev/stdin", "/etc/containerd/certs.d/"+Registry+":5000/hosts.toml")
		return err
	}); err != nil {
		return err
	}
	return pushRegistryImages(ctx)
}

func pushRegistryImages(ctx context.Context) error {
	var sources []string
	for source := range registryImages {
		sources = append(sources, source)
	}
	if err := pullMissing(ctx, sources); err != nil {
		return err
	}
	var err error
	for range 15 {
		if _, err = docker(ctx, RegistryPassword, "login", "--username", RegistryUser, "--password-stdin", "localhost:5000"); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}
	for source, target := range registryImages {
		if _, err := docker(ctx, "", "tag", source, "localhost:5000/"+target); err != nil {
			return err
		}
		if _, err := docker(ctx, "", "push", "--quiet", "localhost:5000/"+target); err != nil {
			return err
		}
	}
	return nil
}

func startAddon(ctx context.Context, name, image, dockerfile string, runArgs ...string) error {
	if containerExists(name) {
		_, err := docker(ctx, "", "start", name)
		return err
	}
	if err := buildImage(ctx, image, dockerfile); err != nil {
		return err
	}
	args := append([]string{"run", "-d", "--name", name, "--hostname", name, "--network", "kind"}, runArgs...)
	_, err := docker(ctx, "", append(args, image)...)
	return err
}
