The **reclaim policy** lives on the **PV**, not the PVC, and decides what happens when the claim
goes away:

- `Delete`: the PV and the underlying storage are removed. This is the default for dynamically
  provisioned volumes, copied from the StorageClass.
- `Retain`: the PV stays, in phase `Released`, with the data intact. An admin decides what to do
  with it next.

It's one of the few PV fields you can change on a live volume:

```sh
kubectl -n st-reclaim get pvc orders                      # VOLUME column
kubectl get pv <name> -o custom-columns=NAME:.metadata.name,POLICY:.spec.persistentVolumeReclaimPolicy
kubectl patch pv <name> -p '{"spec":{"persistentVolumeReclaimPolicy":"Retain"}}'
```

Changing the StorageClass's `reclaimPolicy` doesn't help here: it only applies to PVs created after
the change. A `Released` PV is not reused automatically, because it still remembers its old claim
in `spec.claimRef`. Clearing that is how you hand the data to a new PVC.
