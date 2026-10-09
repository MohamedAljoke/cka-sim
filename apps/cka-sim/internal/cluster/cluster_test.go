package cluster

import (
	"strings"
	"testing"

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/yaml"
)

func TestKindConfigHasOneControlPlaneAndTwoWorkers(t *testing.T) {
	config := parse(t, kindConfig("2"))

	roles := map[v1alpha4.NodeRole]int{}
	for _, node := range config.Nodes {
		roles[node.Role]++
	}

	if roles[v1alpha4.ControlPlaneRole] != 1 || roles[v1alpha4.WorkerRole] != 2 {
		t.Errorf("got roles %v, want 1 control-plane and 2 workers", roles)
	}
}

func TestNodesMatchKindConfig(t *testing.T) {
	if got := len(parse(t, kindConfig("2")).Nodes); got != len(Nodes) {
		t.Errorf("kind.yaml has %d nodes, Nodes lists %d", got, len(Nodes))
	}
}

func TestKubeletIsAllowedToRunOnCgroupV1(t *testing.T) {
	config := parse(t, kindConfig("1"))

	if len(config.KubeadmConfigPatches) != 1 || !strings.Contains(config.KubeadmConfigPatches[0], "failCgroupV1: false") {
		t.Errorf("got patches %q, want one that sets failCgroupV1: false", config.KubeadmConfigPatches)
	}
	if len(config.Nodes) != 3 {
		t.Errorf("got %d nodes, want 3", len(config.Nodes))
	}
}

func TestNoPatchesOnCgroupV2(t *testing.T) {
	if patches := parse(t, kindConfig("2")).KubeadmConfigPatches; len(patches) != 0 {
		t.Errorf("got patches %q, want none", patches)
	}
}

func TestNodeImageTagMatchesItsKindestBase(t *testing.T) {
	version := strings.TrimPrefix(NodeImage, "cka-sim/node:")
	if want := "FROM kindest/node:" + version + "\n"; !strings.HasPrefix(nodeDockerfile, want) {
		t.Errorf("node.Dockerfile must start with %q to match NodeImage %s", want, NodeImage)
	}
}

func parse(t *testing.T, raw string) v1alpha4.Cluster {
	t.Helper()
	var config v1alpha4.Cluster
	if err := yaml.UnmarshalStrict([]byte(raw), &config); err != nil {
		t.Fatalf("kind config does not parse: %v\n%s", err, raw)
	}
	return config
}
