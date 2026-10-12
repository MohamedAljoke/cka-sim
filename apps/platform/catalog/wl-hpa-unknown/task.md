---
id: wl-hpa-unknown
title: HPA targets show unknown
domain: workloads
host: cka-sim-control-plane
topics: [hpa, autoscaling, resources]
weight: 5
---
## Context

`kubectl -n wl-hpa-unknown get hpa shop` shows `cpu: <unknown>/50%` and the HPA never scales.
metrics-server is running and `kubectl top pods` works.

## Task

Find out why HPA `shop` can't read its metric and fix it, so that it shows the real CPU
utilization of Deployment `shop`.

Do not change HPA `shop`. Each container of `shop` should get a CPU request of `50m`.
