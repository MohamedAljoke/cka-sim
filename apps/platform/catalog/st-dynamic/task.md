---
id: st-dynamic
title: Dynamic provisioning end to end
domain: storage
host: cka-sim-control-plane
topics: [pvc, dynamic-provisioning, reclaim-policy]
weight: 4
---
## Task

In Namespace `st-dynamic`:

1. Create a PVC `data` (`200Mi`, `ReadWriteOnce`, StorageClass `standard`), and a Pod `writer`
   (image `busybox:1.36`) that mounts it at `/data` and runs
   `sh -c 'echo hello > /data/hello; sleep 1d'`.
2. Once the PVC is bound, write the name of the PV that was created for it to
   `/opt/course/st-dynamic/pv.txt`, and that PV's reclaim policy to
   `/opt/course/st-dynamic/policy.txt`.
3. Delete Pod `writer` and PVC `data`, and make sure the PV is gone with them.
