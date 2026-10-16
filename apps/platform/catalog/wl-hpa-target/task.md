---
id: wl-hpa-target
title: HPA fails to get the scale
domain: workloads
host: cka-sim-control-plane
topics: [hpa, autoscaling]
weight: 3
---
## Context

HPA `orders` in Namespace `wl-hpa-target` was meant to scale Deployment `orders`, but it never
does anything.

## Task

Find out why and fix HPA `orders`, so that it scales Deployment `orders` and shows its CPU
utilization. Keep its replica bounds and its CPU target.
