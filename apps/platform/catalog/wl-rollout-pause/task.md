---
id: wl-rollout-pause
title: Batch two changes in one rollout
domain: workloads
host: cka-sim-control-plane
topics: [deployments, rollouts, change-cause]
weight: 4
---
## Context

Deployment `web` in Namespace `wl-rollout-pause` runs `nginx:1.26-alpine` and has no resource
requests.

## Task

Make two changes to `web`, but roll them out **together as one new revision**:

1. Image `nginx:1.27-alpine`.
2. A container request of `cpu: 50m` and `memory: 32Mi`.

Pause the rollout before the changes and resume it afterwards. Record the change cause
`upgrade to 1.27 with requests` in the `kubernetes.io/change-cause` annotation, so that it shows
in `kubectl rollout history`.
