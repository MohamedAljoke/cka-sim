---
id: st-storageclass
title: New default StorageClass
domain: storage
host: cka-sim-control-plane
topics: [storageclasses, dynamic-provisioning, default-class]
weight: 4
order: last
---
## Task

1. Create a StorageClass `fast-local` that uses the provisioner `rancher.io/local-path`, keeps
   volumes when their claim is deleted, and binds a volume only once a Pod uses its claim.
2. Make `fast-local` the cluster's **only** default StorageClass, so that a new PVC without a
   `storageClassName` gets `fast-local`.
