The scheduler only places a Pod on a node whose labels match the Pod's rules. `nodeSelector` is the
simple form: every key/value must match exactly. Node affinity is the expressive form: operators
(`In`, `NotIn`, `Exists`, `DoesNotExist`, `Gt`, `Lt`), and two strengths:

- `requiredDuringSchedulingIgnoredDuringExecution`: a hard rule. No matching node means the Pod stays Pending.
- `preferredDuringSchedulingIgnoredDuringExecution`: a weighted wish. The scheduler tries, then places the Pod anywhere.

"IgnoredDuringExecution" means removing the label later doesn't evict running Pods.

```sh
kubectl label node cka-sim-worker2 disktype=ssd
kubectl -n wl-node-affinity run fast --image=nginx:1.27 --dry-run=client -o yaml > fast.yaml
# add under spec:   nodeSelector: {disktype: ssd}
kubectl -n wl-node-affinity create deployment cache --image=nginx:1.27 --replicas=2 --dry-run=client -o yaml > cache.yaml
# add under spec.template.spec: the affinity block from the solution
kubectl apply -f fast.yaml -f cache.yaml
kubectl -n wl-node-affinity get pods -o wide
```

The trap is nesting: `affinity` belongs in the **Pod template** (`spec.template.spec`), not the
Deployment's `spec`. Several `matchExpressions` in one term are ANDed; several `nodeSelectorTerms` are
ORed. Copy the block from the docs page "Assign Pods to Nodes using Node Affinity" rather than typing it.
`kubectl get pods -o wide` shows the node each Pod landed on.
