---
id: nw-cni-bindir
title: Runtime can't find the CNI plugins
domain: networking
host: cka-sim-worker2
topics: [cni, containerd, troubleshooting]
weight: 7
order: last
---
## Context

After a config change on node `cka-sim-worker2`, every new Pod there stays in `ContainerCreating`.
Pods that were already running are fine. Deployment `web` in Namespace `nw-cni-bindir` must run
on that node.

## Task

Find out why and fix the node, so that the Pod of Deployment `web` runs. Do not modify Deployment
`web`, and don't move or copy CNI plugin binaries.
