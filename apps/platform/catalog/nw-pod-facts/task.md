---
id: nw-pod-facts
title: Where a Pod lives on the network
domain: networking
host: cka-sim-control-plane
topics: [pods, ipam, nodes]
weight: 2
---
## Task

For Pod `tracker` in Namespace `nw-pod-facts`, write these lines to
`/opt/course/nw-pod-facts/pod.txt`:

```
ip: <the Pod's IP>
node: <the node it runs on>
podCIDR: <the Pod CIDR assigned to that node>
```
