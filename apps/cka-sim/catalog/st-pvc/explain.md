**Concept.** Storage is split in two so apps don't need to know about disks:
a **PersistentVolume** is a piece of storage the cluster has (cluster-scoped, no namespace);
a **PersistentVolumeClaim** is an app's request for storage (namespaced). The control plane
**binds** a claim to a volume whose storageClassName matches, whose access modes include the
requested ones and whose capacity is at least the request — a 500Mi claim happily binds a 1Gi PV.

**Why `storageClassName: manual` matters.** Leave it off the claim and the cluster's *default*
StorageClass (`standard` in this cluster) is filled in — the claim then gets a dynamically
provisioned volume and never binds your PV. Matching class names is how you pin a claim to a
hand-made volume.

**Reclaim policy `Retain`** keeps the PV and its data when the claim is deleted (the PV goes
`Released` and must be cleaned up by hand); `Delete` removes it.

There is no `kubectl create pv` — copy the PV and PVC skeletons from kubernetes.io/docs
("Configure a Pod to Use a PersistentVolume for Storage").
