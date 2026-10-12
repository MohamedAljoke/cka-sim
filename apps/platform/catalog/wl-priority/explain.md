A **PriorityClass** maps a name to an integer. Pods that reference it get that priority in
`spec.priority`. Pods without one get the global default, or 0. Priority does two things:

1. **Queue order:** Pending Pods with higher priority are scheduled first.
2. **Preemption:** if a high-priority Pod fits nowhere, the scheduler evicts lower-priority Pods
   from a node to make room, then places the Pod there.

```sh
kubectl create priorityclass critical --value=1000000 --description="business-critical"
# add  priorityClassName: critical  under spec: in important.yaml
kubectl apply -f /opt/course/wl-priority/important.yaml
kubectl -n wl-priority get pods -o wide -w          # filler's Pod is preempted, important runs
kubectl -n wl-priority get events | grep -i preempt
```

`globalDefault: true` would make the class the default for every new Pod without one, which is
rarely wanted. `preemptionPolicy: Never` keeps the queue priority but never evicts anything.

Built-in classes `system-cluster-critical` and `system-node-critical` (2 billion and more) protect
CoreDNS, kube-proxy and the CNI. User classes are limited to 1 billion. The evicted `filler` Pod
isn't lost: its Deployment recreates it, and it waits as `Pending` until room frees up.
