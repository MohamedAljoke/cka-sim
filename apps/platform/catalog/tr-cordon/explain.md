`kubectl cordon` sets `spec.unschedulable: true` on a Node. The scheduler then places no new Pods
there (the node also gets the `node.kubernetes.io/unschedulable:NoSchedule` taint), but Pods
already running on it keep running. `kubectl get nodes` shows the node as `Ready,SchedulingDisabled`.

`agent` uses required pod anti-affinity on `kubernetes.io/hostname`, so its two Pods need two
different nodes. The control plane is tainted and the other worker already has one Pod, so the
second Pod has nowhere to go while `cka-sim-worker2` is cordoned. The Pod's events say so:

```sh
kubectl -n tr-cordon describe pod -l app=agent | grep -A3 Events
kubectl get nodes
kubectl uncordon cka-sim-worker2
echo cka-sim-worker2 > /opt/course/tr-cordon/node
```

Uncordoning is enough: the Pending Pod is retried by the scheduler on its own, so nothing needs
a restart. The trap is "fixing" the Deployment (dropping the anti-affinity or a replica), which
hides the symptom and leaves the node out of service.

`kubectl drain` is cordon plus evicting the Pods; `uncordon` undoes both kinds of maintenance.
