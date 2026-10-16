---
id: wl-vpa
title: Ask the VPA for a recommendation
domain: workloads
host: cka-sim-control-plane
topics: [vpa, autoscaling, resources]
weight: 4
---
## Context

Deployment `api` in Namespace `wl-vpa` requests `100m` CPU and `64Mi` memory. The cluster has the
Vertical Pod Autoscaler's CRDs and its recommender installed. Check with
`kubectl get crd | grep verticalpodautoscaler`.

## Task

1. Create a VerticalPodAutoscaler `api` in Namespace `wl-vpa` for Deployment `api`, in
   **recommendation-only** mode: it must never change or restart the Pods.
2. Once it shows a recommendation, write the **target** CPU and memory it recommends for container
   `api` to `/opt/course/wl-vpa/target.txt`, as one line: `cpu=<value> memory=<value>`, for example
   `cpu=50m memory=128Mi`.
