---
id: st-pvc
title: PersistentVolume, claim and Pod
domain: storage
weight: 5
cluster: cka7491
host: cka7491-control-plane
---
## Context

An application in Namespace `storage` needs a volume backed by a directory on the node.

## Task

1. Create a PersistentVolume named `data-pv` with capacity `1Gi`, access mode `ReadWriteOnce`,
   reclaim policy `Retain`, storageClassName `manual`, using hostPath `/mnt/data`.
2. Create a PersistentVolumeClaim named `data-pvc` in Namespace `storage` that requests `500Mi`
   with storageClassName `manual`. It must bind to `data-pv`.
3. Create a Pod named `writer` in Namespace `storage` with image `busybox:1.36` that mounts the
   claim at `/data` and runs `sh -c 'echo ok > /data/ok && sleep 3600'`.
