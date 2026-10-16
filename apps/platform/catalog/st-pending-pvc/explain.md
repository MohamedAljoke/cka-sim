A PVC binds to a PV only if **all** of these match:

| PVC asks for | PV must have |
|---|---|
| `storageClassName` | the same class name (`""` matches only PVs with no class) |
| `accessModes` | every requested mode |
| `resources.requests.storage` | at least that much capacity |
| `selector` (if any) | matching labels |
| `volumeName` (if any) | that exact name |

Here the claim wanted `2Gi` and `ReadWriteMany`, but the PV offers `1Gi` and `ReadWriteOnce`.

```sh
kubectl -n st-pending-pvc describe pvc reports      # no persistent volumes available for this claim
kubectl get pv pv-reports -o wide
```

Almost all of a PVC's spec is **immutable** once created. You can only grow the storage request,
and only on an expandable class. So the fix is to delete and recreate the claim, which means
deleting the Pod that uses it first: the `pvc-protection` finalizer holds the PVC in
`Terminating` while a Pod references it. Save the Pod with `-o yaml` before deleting so you can
recreate it unchanged.

A PVC may ask for less than the PV offers (`500Mi` on a `1Gi` PV binds, and the claim gets all
1Gi), but never more.
