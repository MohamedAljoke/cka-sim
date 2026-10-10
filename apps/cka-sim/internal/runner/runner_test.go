package runner

import (
	"context"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

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

func TestNodeSendsDockerToItsHost(t *testing.T) {
	if env := (Node{}).env(); env != nil {
		t.Errorf("env without a host = %q, want docker's default", env)
	}
	env := Node{DockerHost: "tcp://[fdaa::2]:2375"}.env()
	if !slices.Contains(env, "DOCKER_HOST=tcp://[fdaa::2]:2375") {
		t.Errorf("env = %q, want DOCKER_HOST set", env)
	}
}

func TestNodeFailsWithTheScriptOutput(t *testing.T) {
	n := Node{}

	_, err := n.Run(context.Background(), requireNode(t), "wl-scale", []byte("echo broken >&2\nexit 3\n"))

	if err == nil || !strings.Contains(err.Error(), "broken") {
		t.Errorf("got error %v, want one with the script's output", err)
	}
}

func TestNodeCancelStopsTheScriptInsideTheNode(t *testing.T) {
	host := requireNode(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	began := time.Now()

	_, err := Node{}.Run(ctx, host, "wl-scale", []byte("sleep 4242 &\nsleep 4243\n"))

	if err == nil || !strings.Contains(err.Error(), "stopped") {
		t.Errorf("got error %v, want the script stopped", err)
	}
	if took := time.Since(began); took > 10*time.Second {
		t.Errorf("took %v to stop", took)
	}
	if out, _ := exec.Command("docker", "exec", host, "pgrep", "-f", "sleep 424[23]").Output(); len(out) > 0 {
		t.Errorf("still running in the node: pids %s", out)
	}
}

func requireNode(t *testing.T) string {
	if exec.Command("docker", "inspect", cluster.ControlPlaneNode).Run() != nil {
		t.Skip(cluster.ControlPlaneNode + " is not running; run cka-sim up to include this test")
	}
	return cluster.ControlPlaneNode
}
