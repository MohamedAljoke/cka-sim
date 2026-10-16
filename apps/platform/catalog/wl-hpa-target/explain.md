```sh
kubectl -n wl-hpa-target describe hpa orders
```

```text
AbleToScale  False  FailedGetScale  the HPA controller was unable to get the target's current scale:
             replicasets.apps "orders" not found
```

`scaleTargetRef` names the object whose `/scale` subresource the HPA reads and writes. It says
`kind: ReplicaSet`, and no ReplicaSet is called `orders`: the Deployment's ReplicaSets are named
`orders-<hash>`. Pointing an HPA at a Deployment's ReplicaSet would be wrong anyway, because the
Deployment would undo its changes at the next rollout.

```sh
kubectl -n wl-hpa-target patch hpa orders --type merge -p '{"spec":{"scaleTargetRef":{"kind":"Deployment"}}}'
```

Once the HPA can get the scale, it raises the Deployment to `minReplicas` (2) and starts reading
CPU. `kubectl get hpa` shows the reference as `Deployment/orders`. The same `FailedGetScale` comes
from a typo in the name or a wrong `apiVersion` (for example `v1` instead of `apps/v1`).
