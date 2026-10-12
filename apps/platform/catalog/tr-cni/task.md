---
id: tr-cni
title: Pods stuck in ContainerCreating
domain: troubleshooting
host: cka-sim-worker2
topics: [cni, networking, kubelet, containerd]
weight: 7
order: last
---
## Context

New Pods on node `cka-sim-worker2` stay in `ContainerCreating`. Deployment `web` in Namespace
`tr-cni` must run on that node.

## Task

Find out why and fix the node, so that the Pod of Deployment `web` gets an IP address and runs.

Do not modify Deployment `web`.
