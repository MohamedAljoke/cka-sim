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

func TestOnlyCgroupV1GetsThePatch(t *testing.T) {
	if strings.Contains(kindConfig("2"), "failCgroupV1") {
		t.Error("cgroup v2 got the v1 patch")
	}
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(kindConfig("1")), &parsed); err != nil || parsed["kubeadmConfigPatches"] == nil {
		t.Errorf("cgroup v1 config: %v, %v", parsed, err)
	}
}
