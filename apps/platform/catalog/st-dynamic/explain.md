With **dynamic provisioning** you never create a PV. You create a PVC naming a StorageClass, and
the class's provisioner creates a PV sized and shaped for it, named `pvc-<uid of the PVC>`.
With `WaitForFirstConsumer` (as `standard` uses here), nothing happens until a Pod using the PVC
is scheduled: the PVC stays `Pending` until then, and that's normal.

```sh
kubectl -n st-dynamic get pvc data                 # Pending until writer is scheduled, then Bound
kubectl -n st-dynamic get pvc data -o jsonpath='{.spec.volumeName}'
kubectl get pv <name> -o jsonpath='{.spec.persistentVolumeReclaimPolicy}'
kubectl -n st-dynamic delete pod writer; kubectl -n st-dynamic delete pvc data
kubectl get pv                                      # gone
```

A dynamic PV copies its reclaim policy from the class (`Delete` by default). With `Delete`, removing
the PVC removes the PV **and the data**. The `kubernetes.io/pvc-protection` finalizer keeps a PVC
in `Terminating` while a Pod still uses it, so delete the Pod first. To keep data, switch the PV
to `Retain` before deleting the claim.
