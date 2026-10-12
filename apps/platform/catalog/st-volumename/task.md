---
id: st-volumename
title: Bind a PVC to one chosen PV
domain: storage
host: cka-sim-control-plane
topics: [pv, pvc, binding, volumename]
weight: 3
---
## Context

PVs `pv-orders-a` and `pv-orders-b` are identical (1Gi, `ReadWriteOnce`, StorageClass `orders`),
but only `pv-orders-b` holds last month's data.

## Task

In Namespace `st-volumename`, create PVC `orders` (1Gi, `ReadWriteOnce`, StorageClass `orders`)
that binds to **`pv-orders-b`** and not to `pv-orders-a`. `pv-orders-a` must stay `Available`.
