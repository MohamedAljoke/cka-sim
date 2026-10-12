---
id: wl-hpa-behavior
title: Slow down an HPA's scale-down
domain: workloads
host: cka-sim-control-plane
topics: [hpa, autoscaling]
weight: 4
---
## Context

HPA `worker` in Namespace `wl-hpa-behavior` scales Deployment `worker` down as soon as a busy
minute ends, then up again when the next one starts.

## Task

Change HPA `worker` so that when it scales down:

- it uses the highest recommendation of the last `600` seconds (the stabilization window)
- it removes at most `1` Pod per `60` seconds

Keep its replica bounds, metric and scale-up behaviour as they are.
