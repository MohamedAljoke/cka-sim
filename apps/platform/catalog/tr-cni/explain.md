Before a Pod's containers start, the runtime creates the Pod **sandbox** and its network
namespace. containerd then runs the CNI plugins listed in the first config file in
`/etc/cni/net.d/`, finding each `"type"` as a binary of that name in `/opt/cni/bin/`. If a plugin
binary is missing, sandbox creation fails, the Pod never gets an IP, and it stays
`ContainerCreating` with `FailedCreatePodSandBox` events. Pods that are already running keep
working, which is why only new Pods are affected.

```sh
kubectl -n tr-cni describe pod -l app=web | grep -A2 FailedCreatePodSandBox
ls /etc/cni/net.d/
jq '.plugins[].type' /etc/cni/net.d/10-kindnet.conflist
ls /opt/cni/bin/
vim /etc/cni/net.d/10-kindnet.conflist             # "type": "ptp"
journalctl -u containerd --no-pager | grep -i cni | tail
```

There is no need to restart anything: the kubelet keeps retrying the sandbox, and the runtime
reads the CNI config on every attempt. Compare with another node (`cka-sim-worker`) when you're
unsure what a file should look like. Only the Pod `subnet` differs per node.

On this cluster the CNI is **kindnet**. Its DaemonSet writes this file when it starts, so deleting
the kindnet Pod on that node would also have repaired it. On Calico or Cilium the files differ,
but you debug them the same way.
