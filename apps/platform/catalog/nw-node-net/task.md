---
id: nw-node-net
title: A node's own network settings
domain: networking
host: cka-sim-worker2
topics: [nodes, routing, sysctl, kubeadm-prerequisites]
weight: 2
---
## Task

On node `cka-sim-worker2`, write these facts to `/opt/course/nw-node-net/node.txt`:

```
ip: <the node's InternalIP>
gateway: <the node's default gateway>
ip_forward: <the value of the kernel setting net.ipv4.ip_forward>
```
