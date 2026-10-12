A DaemonSet runs one Pod per matching node and adds one when a node joins. Its Pods still respect
**taints**. The control plane carries `node-role.kubernetes.io/control-plane:NoSchedule`, so
without a toleration the DaemonSet skips it.

```sh
kubectl describe node cka-sim-control-plane | grep Taints
kubectl -n wl-daemonset create deployment log-agent --image=busybox:1.36 --dry-run=client -o yaml -- sleep 1d > ds.yaml
# change kind: Deployment -> DaemonSet, delete replicas, strategy and status, add the tolerations
kubectl apply -f ds.yaml
kubectl -n wl-daemonset get ds,pods -o wide
```

There is no `kubectl create daemonset`. Generate a Deployment and edit it. Two ways to tolerate:

```yaml
tolerations:
- key: node-role.kubernetes.io/control-plane   # just the control plane
  operator: Exists
  effect: NoSchedule
# or
- operator: Exists                             # every taint: what log and network agents do
```

Removing the taint from the node "works" too, but then every Pod can land on the control plane,
and the task forbids it.
