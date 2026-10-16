---
id: wl-quota
title: LimitRange and ResourceQuota
domain: workloads
host: cka-sim-control-plane
topics: [limitrange, resourcequota, requests, limits]
weight: 5
---
## Task

Team limits for Namespace `wl-quota`:

1. Create a LimitRange `defaults` so that every container without its own values gets a default
   request of `cpu: 100m, memory: 64Mi` and a default limit of `cpu: 200m, memory: 128Mi`.
2. Create a ResourceQuota `team-quota` that allows at most `4` Pods, `1` CPU of requests in total
   (`requests.cpu`) and `1Gi` of memory limits in total (`limits.memory`).
3. Try to create Pod `greedy` (image `busybox:1.36`, command `sleep 1d`) that requests and is limited to `cpu: 2`.
   It must be rejected. Write the error message to `/opt/course/wl-quota/rejected.txt`.
