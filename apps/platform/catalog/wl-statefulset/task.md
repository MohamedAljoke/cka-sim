---
id: wl-statefulset
title: StatefulSet with stable storage
domain: workloads
host: cka-sim-control-plane
topics: [statefulsets, headless-services, pvc]
weight: 6
---
## Task

In Namespace `wl-statefulset`:

1. Create a headless Service `db` that selects `app=db` on port `80`.
2. Create a StatefulSet `db` with `3` replicas (image `nginx:1.27-alpine`, label `app=db`) that
   uses Service `db` and gives each Pod its own PersistentVolumeClaim from a template named
   `data`: `100Mi`, `ReadWriteOnce`, StorageClass `standard`, mounted at
   `/usr/share/nginx/html`.

All 3 Pods must be Running with their PVCs Bound.
