---
id: st-emptydir-to-pvc
title: Keep data when the Pod is replaced
domain: storage
host: cka-sim-control-plane
topics: [pvc, emptydir, deployments, persistence]
weight: 4
---
## Context

Deployment `notes` in Namespace `st-emptydir-to-pvc` appends a line to `/data/notes.txt` every time
its Pod starts. The file lives in an `emptyDir`, so it's empty again after every Pod replacement.

## Task

1. Create PVC `notes-data` (`100Mi`, `ReadWriteOnce`, StorageClass `standard`).
2. Change Deployment `notes` to keep `/data` on that PVC instead of the `emptyDir`.
3. Delete the Pod of `notes` once, and make sure the new Pod still sees the lines written by the
   previous one.
