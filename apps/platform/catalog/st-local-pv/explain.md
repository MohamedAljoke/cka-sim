A **local** PersistentVolume is a disk or directory on one specific node. Its required
`nodeAffinity` tells the scheduler where the data is, so any Pod that uses it must run on that
node. If the Pod's own constraints (a `nodeSelector`, affinity, taints) point elsewhere, no node
satisfies both, and the Pod stays `Pending`:

```sh
kubectl -n st-local-pv describe pod app | grep -A3 Events
#   0/3 nodes are available: 1 node(s) didn't match Pod's node affinity/selector,
#   1 node(s) had volume node affinity conflict, 1 node(s) had untolerated taint ...
kubectl get pv local-data -o jsonpath='{.spec.nodeAffinity}'
```

The StorageClass uses `volumeBindingMode: WaitForFirstConsumer` on purpose. The PVC doesn't bind
until a Pod is scheduled, so the scheduler can consider the volume's node and the Pod's
constraints together. With `Immediate`, the claim could bind to a volume on a node the Pod can never
reach.

Local volumes are fast and simple, but the data can't follow the Pod to another node, and if the
node dies, the Pod can't run anywhere else. They suit databases that replicate on their own.
