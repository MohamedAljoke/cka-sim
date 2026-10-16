---
id: st-readonly-share
title: Share a volume read-only
domain: storage
host: cka-sim-control-plane
topics: [pvc, subpath, readonly, volumes]
weight: 4
---
## Context

Namespace `st-readonly-share` has PVC `shared` (`ReadWriteOnce`). Pod `writer` should write its
logs to the sub-folder `writer-logs` of that volume, and an auditor must read everything without
being able to change anything.

## Task

1. Create Pod `writer` (image `busybox:1.36`, command `sh -c 'while true; do date >> /logs/app.log; sleep 2; done'`)
   that mounts **only the sub-folder `writer-logs`** of PVC `shared` at `/logs`, read-write.
2. Create Pod `auditor` (image `busybox:1.36`, command `sleep 1d`) that mounts the **whole** PVC
   `shared` at `/audit`, **read-only**.

From `auditor`, `/audit/writer-logs/app.log` must be readable, and writing anywhere under `/audit`
must fail.
