package kind

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const (
	applyAttempts = 5

	registryHost = "cka-sim-registry"
	chartsHost   = "cka-sim-charts"

	registryUser     = "candidate"
	registryPassword = "cka-registry"

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
	craneImage         = "gcr.io/go-containerregistry/crane@" +
		"sha256:e78770b31258a3846f878036d9c1f63fbe4c871f9f56990bf77fd95c013e3c1b"
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

const registryHosts = `server = "http://` + registryHost + `:5000"

[host."http://` + registryHost + `:5000"]
  capabilities = ["pull", "resolve"]
`

func (p *Provisioner) installAddons(ctx context.Context) error {
	nodes, err := p.Nodes(ctx)
	if err != nil {
		return err
	}

	return errors.Join(
		p.applyAddon(ctx, nodes, metricsServerManifest, metricsServerPath, metricsServerImage),
		p.applyAddon(ctx, nodes, kubeRouterManifest, kubeRouterPath, kubeRouterImage),
		rejectServicesWithoutEndpoints(ctx, nodes),
		p.installTraefik(ctx, nodes),
		p.applyAddon(ctx, nodes, vpaManifest, vpaPath, vpaImage),
		p.stageWebOperator(ctx, nodes),
		installHelm(ctx, nodes),
		p.startRegistry(ctx, nodes),
		p.startAddon(ctx, p.charts(), chartsHost, chartsImage, chartsDockerfile),
	)
}

func (p *Provisioner) registry() string { return p.name + "-registry" }

func (p *Provisioner) charts() string { return p.name + "-charts" }

func (p *Provisioner) controlPlane() string { return p.name + "-control-plane" }

// kube-router's FORWARD and OUTPUT chains accept Pod traffic before kube-proxy's REJECT for a Service
// with no endpoints runs, so that traffic would leave the node instead of being refused.
func rejectServicesWithoutEndpoints(ctx context.Context, nodes []string) error {
	return eachAtOnce(nodes, func(node string) error {
		_, err := docker(ctx, serviceRejectRules, "exec", "-i", node, "nft", "-f", "-")
		return err
	})
}

func (p *Provisioner) installTraefik(ctx context.Context, nodes []string) error {
	if err := p.applyAddon(ctx, nodes, gatewayAPIManifest, gatewayAPIPath); err != nil {
		return err
	}

	if _, err := p.kubectl(ctx, "wait", "--for=condition=Established", "--timeout=60s",
		"crd/gatewayclasses.gateway.networking.k8s.io"); err != nil {
		return err
	}

	return p.applyAddon(ctx, nodes, traefikManifest, traefikPath, traefikImage)
}

func (p *Provisioner) stageWebOperator(ctx context.Context, nodes []string) error {
	if err := buildImage(ctx, webOperatorImage, webOperatorDockerfile); err != nil {
		return err
	}

	if err := loadImages(ctx, nodes, []string{webOperatorImage}); err != nil {
		return err
	}

	_, err := docker(ctx, webOperatorManifest, "exec", "-i", p.controlPlane(),
		"install", "-D", "-m", "644", "/dev/stdin", webOperatorPath)

	return err
}

func (p *Provisioner) applyAddon(ctx context.Context, nodes []string, manifest, path string, images ...string) error {
	if len(images) > 0 {
		if err := loadImages(ctx, nodes, images); err != nil {
			return err
		}
	}

	if _, err := docker(ctx, manifest, "exec", "-i", p.controlPlane(),
		"install", "-D", "-m", "644", "/dev/stdin", path); err != nil {
		return err
	}

	// A fresh cluster's etcd can time out on large manifests such as the Gateway API CRDs; the apply is idempotent.
	var err error

	for range applyAttempts {
		if _, err = p.kubectl(ctx, "apply", "--server-side", "--force-conflicts", "-f", path); err == nil {
			return nil
		}

		time.Sleep(3 * time.Second)
	}

	return err
}

func (p *Provisioner) kubectl(ctx context.Context, args ...string) (string, error) {
	return docker(ctx, "", append([]string{"exec", p.controlPlane(),
		"kubectl", "--kubeconfig", "/etc/kubernetes/admin.conf"}, args...)...)
}

func installHelm(ctx context.Context, nodes []string) error {
	if err := pullMissing(ctx, []string{helmImage}); err != nil {
		return err
	}

	// Only stdout: a warning Docker prints on stderr would end up inside the binary.
	//nolint:gosec // no shell runs the arguments, and the image is the adapter's own constant
	helm, err := exec.CommandContext(ctx, "docker", "run", "--rm", "--entrypoint", "cat",
		helmImage, "/usr/bin/helm").Output()
	if err != nil {
		return fmt.Errorf("copy helm out of %s: %w", helmImage, err)
	}

	return eachAtOnce(nodes, func(node string) error {
		if _, err := docker(ctx, string(helm), "exec", "-i", node,
			"install", "-m", "755", "/dev/stdin", "/usr/local/bin/helm"); err != nil {
			return err
		}

		_, err := docker(ctx, "", "exec", node, "sh", "-c", "helm completion bash > /etc/bash_completion.d/helm")

		return err
	})
}

func (p *Provisioner) startRegistry(ctx context.Context, nodes []string) error {
	if err := p.startAddon(ctx, p.registry(), registryHost, registryImage, registryDockerfile); err != nil {
		return err
	}

	if err := eachAtOnce(nodes, func(node string) error {
		_, err := docker(ctx, registryHosts, "exec", "-i", node,
			"install", "-D", "-m", "644", "/dev/stdin", "/etc/containerd/certs.d/"+registryHost+":5000/hosts.toml")

		return err
	}); err != nil {
		return err
	}

	return pushRegistryImages(ctx)
}

// Pushed from inside the kind network: Docker Desktop's daemon runs in a VM that can't reach a port
// published on this machine's localhost.
func pushRegistryImages(ctx context.Context) error {
	sources := make([]string, 0, len(registryImages))
	for source := range registryImages {
		sources = append(sources, source)
	}

	if err := pullMissing(ctx, append(sources, craneImage)); err != nil {
		return err
	}

	for source, target := range registryImages {
		push := `cat > /tmp/image.tar && crane auth login ` + registryHost + `:5000 -u ` + registryUser +
			` -p ` + registryPassword + ` && crane push --insecure /tmp/image.tar ` + registryHost + `:5000/` + target

		var err error

		for range 15 {
			//nolint:gosec // no shell runs the arguments, and the script is built from the adapter's own constants
			crane := exec.CommandContext(ctx, "docker", "run", "-i", "--rm", "--network", "kind",
				"--entrypoint", "sh", craneImage, "-c", push)
			if err = saveInto(ctx, []string{source}, crane); err == nil {
				break
			}

			time.Sleep(time.Second)
		}

		if err != nil {
			return fmt.Errorf("push %s to the registry: %w", target, err)
		}
	}

	return nil
}

// The task scripts reach the container by host, whatever the cluster is called.
func (p *Provisioner) startAddon(ctx context.Context, name, host, image, dockerfile string, runArgs ...string) error {
	if err := buildImage(ctx, image, dockerfile); err != nil {
		return err
	}

	args := append([]string{"run", "-d", "--name", name, "--hostname", host,
		"--network", "kind", "--network-alias", host}, runArgs...)
	_, err := docker(ctx, "", append(args, image)...)

	return err
}
