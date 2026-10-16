---
id: wl-pdb
title: PodDisruptionBudget for a Deployment
domain: workloads
host: cka-sim-control-plane
topics: [poddisruptionbudgets, drain, availability]
weight: 4
---
## Context

Deployment `api` in Namespace `wl-pdb` runs 3 replicas. During node maintenance it must never drop
below 2 running Pods.

## Task

Create a PodDisruptionBudget `api-pdb` that keeps at least `2` Pods of `api` available during
voluntary disruptions such as `kubectl drain`.
