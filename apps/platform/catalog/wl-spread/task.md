---
id: wl-spread
title: Spread replicas across nodes
domain: workloads
host: cka-sim-control-plane
topics: [topology-spread, scheduling, affinity]
weight: 5
---
## Context

All 4 Pods of Deployment `web` in Namespace `wl-spread` run on the same worker node. If it fails,
the app is down.

## Task

Change Deployment `web` so that its 4 replicas are spread evenly over the worker nodes
`cka-sim-worker` and `cka-sim-worker2`, 2 on each, and stay that way when Pods are recreated.
Use `topologySpreadConstraints` with `maxSkew: 1` on the hostname, and don't allow scheduling
that breaks it.

All 4 Pods must be Running.
