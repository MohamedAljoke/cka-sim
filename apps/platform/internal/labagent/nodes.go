package labagent

// In the order of the lab's kind.yaml.
const ControlPlaneNode = "cka-sim-control-plane"

var Nodes = []string{ControlPlaneNode, "cka-sim-worker", "cka-sim-worker2"}

const BaseHost = "cka-sim-base"
