---
id: st-two-defaults
title: Two default StorageClasses
domain: storage
host: cka-sim-control-plane
topics: [storageclasses, default-class, pvc]
weight: 4
order: last
---
## Context

Someone added StorageClass `legacy` and marked it as default, so the cluster now has two default
classes. New PVCs without a class end up in `legacy`.

## Task

1. Make `standard` the cluster's **only** default StorageClass. Keep `legacy`.
2. Write these two lines to `/opt/course/st-two-defaults/classes.txt`, saying which StorageClass a
   new PVC ends up with in each case (`none` if it gets no class):

   ```
   omitted: <class of a PVC with no storageClassName field>
   empty: <class of a PVC with storageClassName: "">
   ```
