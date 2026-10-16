`Pending` doesn't always mean broken. With a StorageClass in
`volumeBindingMode: WaitForFirstConsumer`, a PVC deliberately stays `Pending` until a Pod that uses
it is **scheduled**. Only then is the volume provisioned or bound, on a node where that Pod can run.

```sh
kubectl -n st-wffc describe pvc cache | tail -3
#   Normal  WaitForFirstConsumer  waiting for first consumer to be created before binding
kubectl get sc standard -o jsonpath='{.volumeBindingMode}'
```

Why wait? Node-local storage (local-path, local PVs) and zonal cloud disks exist in one place. If
the volume were created first, in a random zone or on a random node, the Pod might then be unable to
run anywhere near it. Waiting lets the scheduler choose the node first, honouring the Pod's
affinity, taints and resources, and create the volume there.

The real failures look different: `ProvisioningFailed` events, or `no persistent volumes available
for this claim` with a static class.
