---
id: nw-loadbalancer
title: LoadBalancer without a cloud
domain: networking
host: cka-sim-control-plane
topics: [services, loadbalancer, nodeport]
weight: 3
---
## Context

Deployment `shop` in Namespace `nw-loadbalancer` runs nginx on port `80`. This cluster has no
cloud provider.

## Task

1. Expose Deployment `shop` with a Service `shop` of type `LoadBalancer` on port `80`.
2. Write what `kubectl get service shop` shows as its `EXTERNAL-IP` to
   `/opt/course/nw-loadbalancer/external-ip.txt`.
3. The Service still works through a port on every node. Write that port number to
   `/opt/course/nw-loadbalancer/node-port.txt`.
