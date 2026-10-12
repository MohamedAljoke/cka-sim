A **PodDisruptionBudget** limits *voluntary* disruptions, meaning evictions through the Eviction
API: `kubectl drain`, cluster autoscalers, node upgrades. It can't stop involuntary ones like a
node crash or an OOM kill.

```sh
kubectl -n wl-pdb create pdb api-pdb --selector=app=api --min-available=2
kubectl -n wl-pdb get pdb                # MIN AVAILABLE 2, ALLOWED DISRUPTIONS 1
```

Use `minAvailable` (an absolute number or a percentage) or `maxUnavailable`, not both. With 3
Pods and `minAvailable: 2`, a drain may evict one Pod. It then waits, retrying, until the
Deployment has a replacement Ready somewhere else, and only then evicts the next.

The traps:
- The selector must match the Pods' labels. `ALLOWED DISRUPTIONS 0` with `expectedPods 0` means
  it matches nothing.
- `minAvailable` equal to the replica count (or `maxUnavailable: 0`) blocks every drain forever.
  `kubectl drain` then hangs with `Cannot evict pod as it would violate the pod's disruption budget`.
- Deleting Pods directly (`kubectl delete pod`) ignores PDBs. Only evictions respect them.
