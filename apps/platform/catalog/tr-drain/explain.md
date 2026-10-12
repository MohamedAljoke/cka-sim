`kubectl drain` does two things: it **cordons** the node (`spec.unschedulable: true`) and then
**evicts** every Pod on it, respecting PodDisruptionBudgets. Controllers recreate the evicted Pods
on other nodes. By default it refuses to go on in three cases, and each one has a flag:

| Refusal | Flag | Why it's there |
|---|---|---|
| DaemonSet Pods | `--ignore-daemonsets` | Their controller would put them straight back; they stay. |
| Pods with an `emptyDir` | `--delete-emptydir-data` | That data is lost with the Pod. |
| Pods with no controller | `--force` | Nothing will recreate them; they are gone for good. |

```sh
kubectl drain cka-sim-worker2 --ignore-daemonsets --delete-emptydir-data --force
kubectl get pods -A -o wide --field-selector spec.nodeName=cka-sim-worker2
kubectl get node cka-sim-worker2                   # Ready,SchedulingDisabled
```

After the maintenance, `kubectl uncordon cka-sim-worker2` returns the node to service. Existing
Pods don't move back on their own; only new Pods land there.

The trap is deleting Pods by hand, or `kubectl cordon` alone: a cordon keeps new Pods away but
leaves the running ones where they are. `--force` deletes the standalone `debug` Pod for good,
which is why drain asks you to be explicit.
