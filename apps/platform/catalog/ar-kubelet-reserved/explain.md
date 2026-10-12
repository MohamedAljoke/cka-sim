A node's **Capacity** is all of its CPU and memory. **Allocatable**, what the scheduler may hand
to Pods, is smaller:

```
Allocatable = Capacity - kubeReserved - systemReserved - hard eviction threshold
```

- `systemReserved` keeps room for the OS (sshd, journald, containerd...).
- `kubeReserved` does the same for Kubernetes daemons such as the kubelet.
- `evictionHard` makes the kubelet start evicting Pods **before** the kernel's OOM killer has to
  step in. `memory.available<300Mi` means "keep 300Mi free".

```yaml
# /var/lib/kubelet/config.yaml
evictionHard:
  memory.available: 300Mi
  imagefs.available: 0%       # keep the ones already there: listing evictionHard
  nodefs.available: 0%        # replaces the whole default map
  nodefs.inodesFree: 0%
systemReserved:
  cpu: 500m
  memory: 512Mi
```

```sh
systemctl restart kubelet
kubectl describe node cka-sim-worker2 | grep -A6 -E 'Capacity|Allocatable'
kubectl get --raw /api/v1/nodes/cka-sim-worker2/proxy/configz | jq .kubeletconfig.evictionHard   # what the kubelet really runs with
```

The trap is the eviction map: setting `evictionHard` with only your new key drops every other
threshold that was there. The `configz` endpoint shows the config the kubelet actually loaded,
which proves the restart happened.
