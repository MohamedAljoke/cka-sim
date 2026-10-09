package runner

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/cluster"
)

func TestNodeSourcesLibAndSetsTaskID(t *testing.T) {
	n := Node{Lib: []byte(`greet() { echo "hello $TASK_ID"; }`)}

	out, err := n.Run(context.Background(), requireNode(t), "wl-scale", []byte("greet\n"))

	if err != nil {
		t.Fatal(err)
	}
	if out != "hello wl-scale\n" {
		t.Errorf("output = %q", out)
	}
}

func TestNodeFailsWithTheScriptOutput(t *testing.T) {
	n := Node{}

	_, err := n.Run(context.Background(), requireNode(t), "wl-scale", []byte("echo broken >&2\nexit 3\n"))

	if err == nil || !strings.Contains(err.Error(), "broken") {
		t.Errorf("got error %v, want one with the script's output", err)
	}
}

func requireNode(t *testing.T) string {
	if exec.Command("docker", "inspect", cluster.ControlPlaneNode).Run() != nil {
		t.Skip(cluster.ControlPlaneNode + " is not running; run cka-sim up to include this test")
	}
	return cluster.ControlPlaneNode
}
