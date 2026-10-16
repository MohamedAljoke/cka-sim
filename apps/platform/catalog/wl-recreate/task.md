---
id: wl-recreate
title: Recreate instead of rolling update
domain: workloads
host: cka-sim-control-plane
topics: [deployments, strategy, rollouts]
weight: 3
---
## Context

Deployment `db-migrator` in Namespace `wl-recreate` must never run two versions at the same time,
because both would change the same database.

## Task

1. Change the update strategy of `db-migrator` so that all old Pods are stopped before any new Pod
   starts.
2. Then update its image to `busybox:1.36` with the command unchanged, and make sure the rollout
   finishes.
