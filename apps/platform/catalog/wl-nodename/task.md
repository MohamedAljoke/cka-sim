---
id: wl-nodename
title: Pods without the scheduler
domain: workloads
host: cka-sim-worker2
topics: [static-pods, nodename, scheduling]
weight: 4
---
## Task

In Namespace `wl-nodename`, run two Pods (image `nginx:1.27-alpine`) on node `cka-sim-worker2`
without the scheduler choosing a node:

1. Pod `pinned`, created through the API, placed on `cka-sim-worker2` directly.
2. A **static Pod** named `static-web`, run by the kubelet of `cka-sim-worker2` from a manifest on
   that node.
