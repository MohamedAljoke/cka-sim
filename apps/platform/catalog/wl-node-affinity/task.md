---
id: wl-node-affinity
title: Pin Pods with nodeSelector and node affinity
domain: workloads
host: cka-sim-control-plane
topics: [scheduling, node-affinity, labels]
weight: 4
---
## Task

Work in Namespace `wl-node-affinity`, with image `nginx:1.27`.

1. Label node `cka-sim-worker2` with `disktype=ssd`.
2. Create a Pod `fast` that may only run on nodes labelled `disktype=ssd`. Use `nodeSelector`.
3. Create a Deployment `cache` with **2** replicas whose Pods may only run on nodes where `disktype`
   is `ssd` or `nvme`. Use a **required** node affinity rule.
