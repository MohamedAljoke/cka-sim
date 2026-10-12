---
id: st-access-modes
title: Pick the access mode
domain: storage
host: cka-sim-control-plane
topics: [pv, pvc, access-modes]
weight: 2
---
## Task

For each scenario, write the PersistentVolume access mode it needs to
`/opt/course/st-access-modes/modes.txt`, one line per scenario as `<number>: <mode>`. Use the full
mode names (`ReadWriteOnce`, `ReadOnlyMany`, `ReadWriteMany`, `ReadWriteOncePod`).

1. A single database Pod writes to the volume. No other Pod in the whole cluster may mount it,
   not even on the same node.
2. Twenty web Pods on many nodes serve the same static files. None of them writes.
3. Upload Pods on several nodes all write user files to the same volume.
4. A volume is mounted read-write by one node at a time; a backup sidecar Pod on that same node
   may use it too.
