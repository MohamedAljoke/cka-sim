---
id: st-node-paths
title: Where a Pod's volumes live on the node
domain: storage
host: cka-sim-control-plane
topics: [emptydir, kubelet, crictl, nodes]
weight: 3
---
## Context

Pod `cache` in Namespace `st-node-paths` runs on node `cka-sim-control-plane` and has an
`emptyDir` volume named `scratch`.

## Task

On the node, find these and write them to `/opt/course/st-node-paths/paths.txt`:

```
emptydir: <the directory on the node that holds volume scratch>
pid: <the host PID of container cache's main process>
```
