StatefulSets have a `/scale` subresource like Deployments, so the same HPA works:

```sh
kubectl -n wl-hpa-sts autoscale statefulset db --min=1 --max=3 --cpu=80%
```

What scaling does to a StatefulSet:

- **Order:** Pods go away from the highest ordinal down, so `db-2` first, then `db-1`, and only
  one at a time unless `podManagementPolicy: Parallel` is set. Scale-up creates the missing
  ordinals from the lowest.
- **PVCs stay:** the claims from `volumeClaimTemplates` are not deleted on scale-down, so all 3
  (`data-db-0` to `data-db-2`) remain. When the HPA scales up again, `db-1` and `db-2` get their
  old data back. `spec.persistentVolumeClaimRetentionPolicy.whenScaled: Delete` changes that and
  deletes the claim of each removed Pod.

An idle `db` will drop to 1 replica, but only after the default 5-minute scale-down stabilization
window. That delay is why the HPA here doesn't shrink it at once.
