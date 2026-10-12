A kubeadm cluster runs its parts in three different ways, and knowing which is which tells you
where to look when one breaks:

| Component | Runs as | Where you fix it |
|---|---|---|
| kubelet | systemd service on every node | `systemctl`, `journalctl -u kubelet`, `/var/lib/kubelet/config.yaml` |
| kube-apiserver, kube-scheduler, kube-controller-manager, etcd | static Pods | files in `/etc/kubernetes/manifests/` on the control plane |
| coredns | Deployment in `kube-system` | `kubectl -n kube-system edit deploy coredns` / ConfigMap `coredns` |
| kube-proxy | DaemonSet in `kube-system` | `kubectl -n kube-system edit ds kube-proxy` / ConfigMap `kube-proxy` |

```sh
systemctl status kubelet
ls /etc/kubernetes/manifests/
kubectl -n kube-system get pods -o wide
kubectl -n kube-system get deploy,ds
kubectl -n kube-system get pod kube-scheduler-cka-sim-control-plane -o jsonpath='{.metadata.ownerReferences[0].kind}'   # Node: a mirror Pod
```

The kubelet has to be a host service because it is what starts all the Pods, including the
static ones. Static Pods show up in `kube-system` with the node name as a suffix. They are
mirrors owned by the Node, so editing or deleting them with kubectl changes nothing.
