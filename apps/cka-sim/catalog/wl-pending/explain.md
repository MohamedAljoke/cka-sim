`Pending` with no node means the scheduler found no node that fits. It says why in a
`FailedScheduling` event on the Pod, one reason per node group:

- `didn't match Pod's node affinity/selector`: `api` requires `zone=moon`, and no node has that label.
- `Insufficient memory`: `batch` requests 1Ti. The scheduler adds up **requests**, not real usage, so a
  request bigger than any node's allocatable memory can never fit.
- `had untolerated taint`: the third common reason, fixed with a toleration or by choosing another node.

```sh
kubectl -n wl-pending get pods
kubectl -n wl-pending describe pod -l app=api | tail     # or: kubectl -n wl-pending get events
kubectl -n wl-pending edit deployment api                # delete the affinity block
kubectl -n wl-pending set resources deployment batch --requests=memory=128Mi
kubectl -n wl-pending get pods -o wide
```

Fix the **Deployment**, not the Pod: the ReplicaSet replaces an edited Pod with a fresh copy of the old
template. Labelling a node `zone=moon` would also schedule `api`, but it changes the cluster to fit a
wrong spec, which this task forbids. `kubectl describe node <name>` shows `Allocatable` and the requests
already placed, which helps size a request.
