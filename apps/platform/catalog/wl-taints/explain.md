A taint on a node repels Pods; a toleration on a Pod lets it ignore that taint. The effect decides how
hard: `NoSchedule` keeps new Pods off, `PreferNoSchedule` only discourages, `NoExecute` also evicts
running Pods that don't tolerate it. The control plane node carries
`node-role.kubernetes.io/control-plane:NoSchedule`, which is why ordinary Pods never land there.

```sh
kubectl taint node cka-sim-worker dedicated=gpu:NoSchedule
kubectl describe node cka-sim-worker | grep -i taints
kubectl -n wl-taints run gpu-job --image=nginx:1.27 --dry-run=client -o yaml > gpu.yaml
# add under spec: the tolerations and nodeSelector from the solution
kubectl apply -f gpu.yaml
kubectl -n wl-taints run regular --image=nginx:1.27
kubectl -n wl-taints get pods -o wide
```

The trap: a toleration only **allows** a Pod onto the node, it never **pulls** it there. `gpu-job` with a
toleration alone may land on `cka-sim-worker2`. To dedicate a node you need both: the taint keeps others
off, and a `nodeSelector` (or node affinity) on the special Pods pulls them on.
Remove a taint with a trailing `-`: `kubectl taint node cka-sim-worker dedicated-`.
