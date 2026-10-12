---
id: nw-fixed-clusterip
title: Service with a fixed ClusterIP
domain: networking
host: cka-sim-control-plane
topics: [services, clusterip, service-cidr]
weight: 3
---
## Context

A legacy app has the address `10.96.200.200` hard-coded for its cache. Deployment `cache` in
Namespace `nw-fixed-clusterip` runs nginx on port `80`.

## Task

1. Create Service `cache` for Deployment `cache` on port `80` whose ClusterIP is exactly
   `10.96.200.200`.
2. Try to create another Service `cache-outside` with the ClusterIP `10.200.0.10`. It will be
   rejected. Write the error message to `/opt/course/nw-fixed-clusterip/error.txt`.
