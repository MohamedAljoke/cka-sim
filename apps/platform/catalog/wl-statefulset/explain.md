A StatefulSet gives each replica an identity that survives rescheduling:

- **Stable names:** `db-0`, `db-1`, `db-2`, created in order and removed in reverse.
- **Stable storage:** one PVC per Pod from `volumeClaimTemplates`, named
  `<template>-<pod>` (`data-db-0`). A recreated `db-1` gets `data-db-1` back.
- **Stable DNS:** with a **headless** Service (`clusterIP: None`) named in `serviceName`, each Pod
  resolves as `<pod>.<service>.<namespace>.svc.cluster.local`.

```sh
kubectl -n wl-statefulset get sts,pods,pvc
kubectl -n wl-statefulset exec db-0 -- nslookup db-1.db.wl-statefulset.svc.cluster.local
kubectl -n wl-statefulset delete pod db-1 && kubectl -n wl-statefulset get pod db-1 -w   # same name, same PVC
```

There is no `kubectl create statefulset`. Start from the docs page "StatefulSets", which has a
complete Service + StatefulSet example.

Scaling down or deleting a StatefulSet **keeps** its PVCs by default, so data isn't lost by
accident (`persistentVolumeClaimRetentionPolicy` changes that). With `WaitForFirstConsumer`,
each PVC binds only once its Pod is scheduled, which is why the Pods come up one after another.
