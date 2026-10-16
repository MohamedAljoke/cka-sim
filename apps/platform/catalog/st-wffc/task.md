---
id: st-wffc
title: PVC waits for its first consumer
domain: storage
host: cka-sim-control-plane
topics: [pvc, storageclasses, volume-binding-mode]
weight: 3
---
## Context

PVC `cache` in Namespace `st-wffc` has been `Pending` for a while. The team thinks the storage is
broken.

## Task

1. Find the event that explains why `cache` is `Pending` and write its message to
   `/opt/course/st-wffc/reason.txt`.
2. Create Pod `cache-user` (image `busybox:1.36`, command `sleep 1d`) that mounts PVC `cache` at
   `/cache`, so that the PVC gets bound.
