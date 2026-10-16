---
id: st-pvc-selector
title: Pick a PV by label
domain: storage
host: cka-sim-control-plane
topics: [pvc, pv, labels, selectors]
weight: 3
---
## Context

The StorageClass `tiered` has two pre-made PVs: `pv-ssd` (label `disk=ssd`, 2Gi) and `pv-hdd`
(label `disk=hdd`, 5Gi).

## Task

1. In Namespace `st-pvc-selector`, create PVC `fast` (request `1Gi`, `ReadWriteOnce`, StorageClass
   `tiered`) that may only bind to a PV labelled `disk=ssd`.
2. Write the storage capacity the bound claim reports in its status to
   `/opt/course/st-pvc-selector/capacity.txt`.
