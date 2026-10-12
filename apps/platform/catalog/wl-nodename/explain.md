Two ways to put a Pod on a node without kube-scheduler:

**`spec.nodeName`.** Scheduling is nothing more than the scheduler filling in `nodeName`. Set
it yourself and the kubelet on that node runs the Pod right away. No filters apply, so taints,
resources and affinity are all skipped, and there is no `Scheduled` event.

```sh
kubectl -n wl-nodename run pinned --image=nginx:1.27-alpine --overrides='{"spec":{"nodeName":"cka-sim-worker2"}}'
```

**Static Pods.** The kubelet runs every manifest in its `staticPodPath` (from
`/var/lib/kubelet/config.yaml`, usually `/etc/kubernetes/manifests`) by itself, even with no API
server. It then creates a read-only **mirror Pod** in the API named `<name>-<node>`.

```sh
grep staticPodPath /var/lib/kubelet/config.yaml
kubectl -n wl-nodename run static-web --image=nginx:1.27-alpine --dry-run=client -o yaml > /etc/kubernetes/manifests/static-web.yaml
kubectl -n wl-nodename get pods -o wide        # static-web-cka-sim-worker2
```

The manifest must be written **on that node**: ssh there first. `kubectl delete pod
static-web-cka-sim-worker2` only deletes the mirror, and the kubelet recreates it. Delete the file
to stop the Pod. This is exactly how kubeadm runs the control plane.
