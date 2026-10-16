---
id: nw-veth
title: Find a Pod's veth on the node
domain: networking
host: cka-sim-worker2
topics: [network-namespaces, veth, crictl, cni]
weight: 5
---
## Context

Pod `probe` in Namespace `nw-veth` runs on node `cka-sim-worker2`.

## Task

Find the host side of Pod `probe`'s network interface pair on `cka-sim-worker2`, and write these
lines to `/opt/course/nw-veth/veth.txt`:

```
pid: <host PID of container probe's main process>
pod-ip: <the IP inside the Pod's network namespace, on eth0>
host-if: <name of the veth interface on the node that is paired with the Pod's eth0>
```
