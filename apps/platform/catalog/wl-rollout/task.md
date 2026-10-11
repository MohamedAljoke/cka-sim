---
id: wl-rollout
title: Roll back a broken rollout
domain: workloads
host: cka-sim-control-plane
topics: [deployments, rollouts, rollback]
weight: 5
---
## Context

The most recent update of Deployment `shop` in Namespace `release` left it unable to start new
Pods.

## Task

1. Roll Deployment `shop` back to the previous revision that ran successfully.
2. Write the revision number the Deployment is on **after** the rollback to
   `/opt/course/wl-rollout/revision.txt`.
3. Change the update strategy of `shop` so that during a rolling update at most `1` Pod is
   unavailable and at most `2` extra Pods are created.
