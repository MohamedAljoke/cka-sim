---
id: nw-clusterip
title: Expose a Deployment inside the cluster
domain: networking
host: cka-sim-control-plane
topics: [services, clusterip, ports]
weight: 3
---
## Context

Deployment `api` in Namespace `nw-clusterip` serves HTTP on container port `8080`.

## Task

Create a ClusterIP Service `api` that makes the Deployment reachable inside the cluster on port
`80` as `http://api.nw-clusterip`.
