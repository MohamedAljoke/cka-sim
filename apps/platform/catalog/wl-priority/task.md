---
id: wl-priority
title: PriorityClass and preemption
domain: workloads
host: cka-sim-control-plane
topics: [priorityclasses, preemption, scheduling]
weight: 5
---
## Context

Node `cka-sim-worker2` is almost full: Deployment `filler` in Namespace `wl-priority` uses most of
its CPU. Pod `important`, described in `/opt/course/wl-priority/important.yaml`, doesn't fit next
to it.

## Task

1. Create a PriorityClass `critical` with value `1000000` that is **not** the cluster default.
2. Run Pod `important` from `/opt/course/wl-priority/important.yaml` with PriorityClass `critical`,
   so that the scheduler makes room for it on `cka-sim-worker2`.

Do not change the CPU request or the node selector of `important`, and do not modify Deployment
`filler`.
