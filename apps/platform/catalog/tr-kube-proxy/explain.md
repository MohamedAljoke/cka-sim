A ClusterIP is virtual: no process listens on it. **kube-proxy**, a DaemonSet with one Pod per
node, watches Services and EndpointSlices and writes iptables (or ipvs or nftables) rules that
send traffic for each ClusterIP to a backend Pod. When kube-proxy stops, the rules it already
wrote stay in place, so old Services keep working while new ones get no rules at all. That's the
sneaky part of this fault.

```sh
kubectl -n tr-kube-proxy get svc,endpointslices        # endpoints exist, so it's not the selector
kubectl -n kube-system get daemonset kube-proxy        # DESIRED 0
kubectl -n kube-system get daemonset kube-proxy -o jsonpath='{.spec.template.spec.nodeSelector}'
kubectl -n kube-system edit daemonset kube-proxy       # remove the kube-proxy: enabled selector
kubectl -n kube-system rollout status daemonset kube-proxy
iptables-save | grep tr-kube-proxy/web                 # on a node: the rules are back
```

A DaemonSet only runs on nodes that match its `nodeSelector` and whose taints it tolerates.
Labelling every node `kube-proxy=enabled` would also work, but it treats the symptom. Remove the
selector that shouldn't be there.

When a Service has endpoints but doesn't answer, check kube-proxy next: its Pods, its logs, and
its mode in `kubectl -n kube-system get configmap kube-proxy -o yaml`.
