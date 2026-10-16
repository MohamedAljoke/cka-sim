---
id: wl-init
title: Wait for a dependency with an init container
domain: workloads
host: cka-sim-control-plane
topics: [init-containers, dns, services]
weight: 4
---
## Context

Deployment `web` in Namespace `wl-init` crashes when it starts before its database Service exists.
Deployment `db` is running, but has no Service yet.

## Task

1. Add an init container `wait-for-db` (image `busybox:1.36`) to Deployment `web` that waits until
   the name `db.wl-init.svc.cluster.local` resolves, and only then lets the main container start.
2. Create Service `db` for Deployment `db` on port `80`.

`web` must end up with 2 Ready Pods.
