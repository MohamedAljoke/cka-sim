---
id: wl-quota-rollout
title: Rollout stuck on a quota
domain: workloads
host: cka-sim-control-plane
topics: [deployments, rollouts, resourcequota, maxsurge]
weight: 5
---
## Context

An image update of Deployment `api` in Namespace `wl-quota-rollout` hangs: `kubectl rollout status`
never finishes and the Deployment reports `ProgressDeadlineExceeded`. The Pods themselves look
fine.

## Task

1. Find the event that explains why the rollout is stuck and write its message to
   `/opt/course/wl-quota-rollout/reason.txt`.
2. Get the rollout to `nginx:1.27-alpine` finished, with 2 Ready replicas.

Do not change the ResourceQuota, the replica count, or the Pods' CPU request.
