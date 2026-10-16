---
id: wl-hpa
title: Autoscale a Deployment on CPU
domain: workloads
host: cka-sim-control-plane
topics: [hpa, autoscaling, metrics-server]
weight: 4
---
## Context

Deployment `api` in Namespace `wl-hpa` runs one Pod. Traffic varies during the day. The cluster
runs metrics-server.

## Task

Create a HorizontalPodAutoscaler (`autoscaling/v2`) named `api` in Namespace `wl-hpa` for
Deployment `api`:

- at least `2` and at most `6` replicas
- a target average CPU utilization of `60%`
