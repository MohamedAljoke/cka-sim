---
id: st-pending-pvc
title: PVC stuck in Pending
domain: storage
host: cka-sim-control-plane
topics: [pvc, pv, binding, troubleshooting]
weight: 5
---
## Context

PersistentVolume `pv-reports` was created for the reports app. Its claim `reports` in Namespace
`st-pending-pvc` stays `Pending`, and so does Pod `reporter`, which uses it.

## Task

Find out why and fix it, so that PVC `reports` binds to PV `pv-reports` and Pod `reporter` runs.

Do not modify PV `pv-reports`.
