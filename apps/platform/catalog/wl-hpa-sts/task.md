---
id: wl-hpa-sts
title: Autoscale a StatefulSet
domain: workloads
host: cka-sim-control-plane
topics: [hpa, statefulsets, pvc]
weight: 5
---
## Context

StatefulSet `db` in Namespace `wl-hpa-sts` runs 3 replicas, each with its own PersistentVolumeClaim.
It is idle most of the day.

## Task

1. Create a HorizontalPodAutoscaler (`autoscaling/v2`) named `db` in Namespace `wl-hpa-sts` for
   StatefulSet `db`: `1` to `3` replicas, target average CPU utilization `80%`.
2. When the HPA has scaled `db` down to 1 replica, how many PersistentVolumeClaims of `db` will
   exist? Write the number to `/opt/course/wl-hpa-sts/pvcs.txt`.
3. Which Pod does the StatefulSet remove first when it scales down from 3? Write its name to
   `/opt/course/wl-hpa-sts/first-removed.txt`.
