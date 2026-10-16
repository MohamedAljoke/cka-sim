---
id: wl-hpa-load
title: Load-test an HPA
domain: workloads
host: cka-sim-control-plane
topics: [hpa, autoscaling, load-testing]
weight: 6
---
## Context

Deployment `web` in Namespace `wl-hpa-load` runs behind Service `web`, with HPA `web` from 1 to 4
replicas. Nobody has seen it scale yet.

## Task

1. Create a Pod named `load` in Namespace `wl-hpa-load` with image `busybox:1.36`. It must
   request `http://web` in an endless loop, so that HPA `web` scales Deployment `web` to at least 2
   replicas.
2. After the load stops, how many seconds does HPA `web` wait before it removes Pods? Write the
   number to `/opt/course/wl-hpa-load/scale-down.txt`.
