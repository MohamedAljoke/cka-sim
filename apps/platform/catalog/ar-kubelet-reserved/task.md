---
id: ar-kubelet-reserved
title: Reserve resources for the node
domain: architecture
host: cka-sim-worker2
topics: [kubelet, eviction, allocatable, resources]
weight: 5
---
## Context

Pods on node `cka-sim-worker2` can use up all of its memory and CPU, leaving nothing for the OS
and starving system daemons.

## Task

Configure the kubelet of `cka-sim-worker2`, in its config file, so that:

1. `500m` CPU and `512Mi` memory are reserved for system daemons (`systemReserved`).
2. Pods are evicted when the node's available memory drops below `300Mi` (a hard eviction
   threshold). Keep the eviction thresholds that are already configured.

Apply the change, and make sure the node is `Ready` and reports the smaller **Allocatable**
resources.
