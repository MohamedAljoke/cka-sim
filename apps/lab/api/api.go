// Package api is how a platform reaches a lab: the agent's client, the shell protocol the page, the platform
// and the agent all speak, and the nodes task scripts run on. It is the only lab package a platform imports.
package api

// The lab's kind nodes, in the order of kind.yaml. A task names one as its host.
const ControlPlaneNode = "cka-sim-control-plane"

var Nodes = []string{ControlPlaneNode, "cka-sim-worker", "cka-sim-worker2"}
