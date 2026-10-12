---
id: ar-cni-facts
title: Which CNI runs here
domain: architecture
host: cka-sim-control-plane
topics: [cni, networking, ipam]
weight: 3
---
## Task

Find out about this cluster's Pod network and write these lines to
`/opt/course/ar-cni-facts/cni.txt`:

```
plugin: <name of the CNI plugin, e.g. calico, cilium, flannel>
config: <full path of the CNI config file on the nodes>
podCIDR: <the cluster-wide Pod CIDR>
```
