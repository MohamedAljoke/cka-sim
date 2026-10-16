---
id: nw-cidrs
title: Pod, Service and node CIDRs
domain: networking
host: cka-sim-control-plane
topics: [ipam, pod-cidr, service-cidr, control-plane]
weight: 3
---
## Task

Write this cluster's address ranges to `/opt/course/nw-cidrs/cidrs.txt`:

```
pods: <the cluster-wide Pod CIDR>
services: <the Service (ClusterIP) CIDR>
cka-sim-worker: <the Pod CIDR assigned to node cka-sim-worker>
```
