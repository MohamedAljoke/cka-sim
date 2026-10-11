---
id: wl-taints
title: Taint a node and tolerate it
domain: workloads
host: cka-sim-control-plane
topics: [scheduling, taints, tolerations]
weight: 4
---
## Context

Node `cka-sim-worker` is to be kept for GPU work.

## Task

Work in Namespace `wl-taints`, with image `nginx:1.27`.

1. Taint node `cka-sim-worker` with `dedicated=gpu:NoSchedule`.
2. Create a Pod `gpu-job` that tolerates this taint and runs on `cka-sim-worker`.
3. Create a Pod `regular` with no tolerations. Make sure it is Running and see where it lands.
