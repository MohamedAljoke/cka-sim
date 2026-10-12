---
id: wl-daemonset
title: DaemonSet on every node
domain: workloads
host: cka-sim-control-plane
topics: [daemonsets, tolerations, taints]
weight: 4
---
## Task

In Namespace `wl-daemonset`, create a DaemonSet `log-agent` (image `busybox:1.36`, command
`sleep 1d`, label `app=log-agent`) that runs exactly one Pod on **every** node of the cluster,
including the control plane.

Don't remove any taint from any node.
