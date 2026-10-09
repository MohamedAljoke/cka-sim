---
id: ar-etcd
title: etcd backup and restore
domain: architecture
host: cka-sim-control-plane
topics: [etcd, backup, restore]
weight: 8
---
## Context

The cluster runs etcd as a static Pod on its control plane node. The etcd client certificates
are in `/etc/kubernetes/pki/etcd/`.

## Task

1. Take a snapshot of the running etcd and save it to `/opt/course/ar-etcd/etcd-snapshot.db`.
2. Someone deleted ConfigMap `marker` from Namespace `etcd-check` by mistake. A snapshot taken
   before that is at `/opt/course/ar-etcd/etcd-backup-old.db`. Restore the cluster from it.

After the restore the cluster must be healthy and ConfigMap `etcd-check/marker` must exist again.
