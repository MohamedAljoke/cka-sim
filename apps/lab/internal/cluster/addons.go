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
	kubeRouterImage    = "cloudnativelabs/kube-router:v2.11.1"
	kubeRouterPath     = "/etc/cka-sim/addons/kube-router.yaml"
	gatewayAPIPath     = "/etc/cka-sim/addons/gateway-api.yaml"
	traefikImage       = "traefik:v3.7.13"
	traefikPath        = "/etc/cka-sim/addons/traefik.yaml"
	vpaImage           = "registry.k8s.io/autoscaling/vpa-recommender:1.8.0"
	vpaPath            = "/etc/cka-sim/addons/vpa.yaml"
	webOperatorImage   = "cka-sim/web-operator:1.0"
	webOperatorPath    = "/etc/cka-sim/addons/web-operator.yaml"
)

//go:embed metrics-server.yaml
var metricsServerManifest string

//go:embed kube-router.yaml
var kubeRouterManifest string

//go:embed gateway-api.yaml
var gatewayAPIManifest string

//go:embed traefik.yaml
var traefikManifest string

//go:embed service-reject.nft
var serviceRejectRules string

//go:embed vpa.yaml
var vpaManifest string

//go:embed web-operator.Dockerfile
var webOperatorDockerfile string

//go:embed web-operator.yaml
var webOperatorManifest string

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
	if err := waitForAPIServer(ctx); err != nil {
		return err
	}
	return errors.Join(
		c.applyAddon(ctx, metricsServerManifest, metricsServerPath, metricsServerImage),
		c.applyAddon(ctx, kubeRouterManifest, kubeRouterPath, kubeRouterImage),
		rejectServicesWithoutEndpoints(ctx),
		c.installTraefik(ctx),
		c.applyAddon(ctx, vpaManifest, vpaPath, vpaImage),
		c.stageWebOperator(ctx),
		installHelm(ctx),
		startRegistry(ctx),
		startAddon(ctx, Charts, chartsImage, chartsDockerfile),
	)
}

// kube-router's FORWARD and OUTPUT chains accept Pod traffic before kube-proxy's REJECT for a Service
// with no endpoints runs, so that traffic would leave the node instead of being refused.
func rejectServicesWithoutEndpoints(ctx context.Context) error {
	return eachAtOnce(Nodes, func(node string) error {
		_, err := docker(ctx, serviceRejectRules, "exec", "-i", node, "nft", "-f", "-")
		return err
	})
}

func (c *Cluster) installTraefik(ctx context.Context) error {
	if err := c.applyAddon(ctx, gatewayAPIManifest, gatewayAPIPath); err != nil {
		return err
	}
	if _, err := kubectl(ctx, "wait", "--for=condition=Established", "--timeout=60s", "crd/gatewayclasses.gateway.networking.k8s.io"); err != nil {
		return err
	}
	return c.applyAddon(ctx, traefikManifest, traefikPath, traefikImage)
}

func (c *Cluster) stageWebOperator(ctx context.Context) error {
	if err := buildImage(ctx, webOperatorImage, webOperatorDockerfile); err != nil {
		return err
	}
	if err := c.LoadImages(ctx, []string{webOperatorImage}); err != nil {
		return err
	}
	_, err := docker(ctx, webOperatorManifest, "exec", "-i", ControlPlaneNode, "install", "-D", "-m", "644", "/dev/stdin", webOperatorPath)
	return err
}

func (c *Cluster) applyAddon(ctx context.Context, manifest, path string, images ...string) error {
	if len(images) > 0 {
		if err := c.LoadImages(ctx, images); err != nil {
			return err
		}
	}
	if _, err := docker(ctx, manifest, "exec", "-i", ControlPlaneNode, "install", "-D", "-m", "644", "/dev/stdin", path); err != nil {
		return err
	}
	_, err := kubectl(ctx, "apply", "--server-side", "--force-conflicts", "-f", path)
	return err
}

func waitForAPIServer(ctx context.Context) error {
	var err error
	for range 60 {
		if _, err = kubectl(ctx, "get", "--raw", "/readyz"); err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return err
}

func kubectl(ctx context.Context, args ...string) (string, error) {
	return docker(ctx, "", append([]string{"exec", ControlPlaneNode, "kubectl", "--kubeconfig", "/etc/kubernetes/admin.conf"}, args...)...)
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
