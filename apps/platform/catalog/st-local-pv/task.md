---
id: st-local-pv
title: Pod Pending on a local volume
domain: storage
host: cka-sim-worker2
topics: [local-volumes, node-affinity, pv, scheduling]
weight: 4
---
## Context

Pod `app` in Namespace `st-local-pv` uses PVC `data`, which can only bind to the local PV
`local-data`. The Pod stays `Pending`.

## Task

Find out why and fix **the Pod**, so that PVC `data` binds to PV `local-data` and Pod `app` runs.
Pod `app` is defined in `/opt/course/st-local-pv/app.yaml`.

Do not modify PV `local-data` or StorageClass `local-storage`.
