package kind

import (
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestStepOfReadsKindsProgressLines(t *testing.T) {
	for _, tc := range []struct{ line, want string }{
		{" • Ensuring node image (kindest/node:v1.37.0) 🖼  ...\n", "Ensuring node image (kindest/node:v1.37.0)"},
		{" • Preparing nodes 📦 📦 📦   ...\n", "Preparing nodes"},
		{" • Starting control-plane 🕹️  ...\n", "Starting control-plane"},
		{" • Waiting ≤ 5m0s for control-plane = Ready ⏳  ...\n", "Waiting ≤ 5m0s for control-plane = Ready"},
	} {
		got, ok := stepOf(tc.line)

		if !ok || got != tc.want {
			t.Errorf("stepOf(%q) = %q, %v; want %q", tc.line, got, ok, tc.want)
		}
	}
}

func TestStepOfSkipsEverythingElse(t *testing.T) {
	for _, line := range []string{" ✓ Preparing nodes 📦\n", "Creating cluster \"cka-sim\" ...\n", ""} {
		if step, ok := stepOf(line); ok {
			t.Errorf("stepOf(%q) = %q, want no step", line, step)
		}
	}
}

func TestTheClusterHasOneControlPlaneAndTwoWorkers(t *testing.T) {
	var parsed struct {
		Nodes []struct{ Role string }
	}
	if err := yaml.Unmarshal([]byte(kindConfig("2")), &parsed); err != nil {
		t.Fatal(err)
	}

	var roles []string
	for _, n := range parsed.Nodes {
		roles = append(roles, n.Role)
	}
	if strings.Join(roles, ",") != "control-plane,worker,worker" {
		t.Errorf("roles = %v", roles)
	}
}

func TestTheNodeImageTagMatchesItsDockerfiles(t *testing.T) {
	_, version, _ := strings.Cut(nodeImage, ":")

	if !strings.HasPrefix(nodeDockerfile, "FROM kindest/node:"+version+"\n") {
		t.Errorf("%s isn't built from kindest/node:%s", nodeImage, version)
	}
	if !strings.HasPrefix(webOperatorDockerfile, "FROM "+nodeImage+" AS node\n") {
		t.Errorf("web-operator.Dockerfile doesn't take kubectl from %s", nodeImage)
	}
}

func TestCgroupV1GetsTheGracePeriodThenItsOwnPatch(t *testing.T) {
	patches := patchesOf(t, kindConfig("1"))

	if len(patches) != 2 || !strings.Contains(patches[0], "node-monitor-grace-period") ||
		!strings.Contains(patches[1], "failCgroupV1: false") {
		t.Errorf("patches = %q, want the grace period's, then failCgroupV1: false", patches)
	}
}

func TestCgroupV2GetsOnlyTheGracePeriod(t *testing.T) {
	patches := patchesOf(t, kindConfig("2"))

	if len(patches) != 1 || !strings.Contains(patches[0], "node-monitor-grace-period") {
		t.Errorf("patches = %q, want only the node grace period", patches)
	}
}

func TestMetricsServerRunsThePreloadedImageWithoutKubeletCerts(t *testing.T) {
	for _, want := range []string{"image: " + metricsServerImage + "\n", "- --kubelet-insecure-tls\n"} {
		if !strings.Contains(metricsServerManifest, want) {
			t.Errorf("metrics-server.yaml lacks %q", want)
		}
	}
}

func patchesOf(t *testing.T, config string) []string {
	t.Helper()

	var parsed struct {
		KubeadmConfigPatches []string `json:"kubeadmConfigPatches"`
	}
	if err := yaml.Unmarshal([]byte(config), &parsed); err != nil {
		t.Fatal(err)
	}

	return parsed.KubeadmConfigPatches
}
