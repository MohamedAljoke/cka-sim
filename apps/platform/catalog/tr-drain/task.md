---
id: tr-drain
title: Drain a node for maintenance
domain: troubleshooting
host: cka-sim-control-plane
topics: [nodes, drain, daemonsets, maintenance]
weight: 5
order: last
---
## Context

Node `cka-sim-worker2` needs a kernel update. Namespace `tr-drain` runs Deployments `api` and
`cache` (with an `emptyDir`), a DaemonSet `node-agent`, and a standalone Pod `debug` on that node.

## Task

Drain node `cka-sim-worker2` for the maintenance and leave it drained:

- No Pods may run on it any more, except Pods of DaemonSets.
- New Pods must not be scheduled to it.
- Deployments `api` and `cache` must keep all their replicas Ready on other nodes.

Do not modify the Deployments.
