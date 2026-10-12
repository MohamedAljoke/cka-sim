---
id: nw-routes
title: How a node routes to other nodes' Pods
domain: networking
host: cka-sim-worker
topics: [routing, cni, pod-cidr, nodes]
weight: 3
---
## Task

On node `cka-sim-worker`, find out how traffic to Pods on node `cka-sim-worker2` leaves this node.
Write these lines to `/opt/course/nw-routes/route.txt`:

```
podCIDR: <the Pod CIDR of cka-sim-worker2>
via: <the next-hop IP this node uses for it>
dev: <the network interface it leaves through>
```
