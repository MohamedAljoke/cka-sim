---
id: wl-hpa-multi
title: Autoscale on CPU and memory
domain: workloads
host: cka-sim-control-plane
topics: [hpa, autoscaling, metrics-server]
weight: 5
---
## Context

Deployment `report` in Namespace `wl-hpa-multi` needs both more CPU and more memory when big
reports run.

## Task

1. Create a HorizontalPodAutoscaler (`autoscaling/v2`) named `report` in Namespace `wl-hpa-multi`
   for Deployment `report`, with `1` to `5` replicas and two metrics:
   - CPU: target average utilization `70%`
   - memory: target average value `100Mi` per Pod
2. When the two metrics ask for different replica counts, which one does the HPA follow? Write
   one word to `/opt/course/wl-hpa-multi/decides.txt`: `cpu`, `memory`, `highest` (the one that
   asks for more replicas) or `lowest` (the one that asks for fewer).
