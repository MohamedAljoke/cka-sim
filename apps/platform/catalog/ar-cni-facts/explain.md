Kubernetes itself doesn't wire Pods together. A **CNI plugin** does: the runtime calls it for every
new Pod sandbox, and it gives the Pod an interface, an IP and routes. Three places tell you which
one a cluster uses:

```sh
ls /etc/cni/net.d/                           # the runtime uses the first file, in lexical order
cat /etc/cni/net.d/10-kindnet.conflist       # plugin chain: ptp + host-local IPAM + portmap
ls /opt/cni/bin/                             # the plugin binaries
kubectl -n kube-system get ds                # the CNI's agent DaemonSet (kindnet, calico-node, cilium...)
```

The **cluster-wide Pod CIDR** is the controller-manager's `--cluster-cidr` (also in `kubectl -n
kube-system get cm kubeadm-config -o yaml` as `podSubnet`). The controller-manager splits it into
one smaller range per node, `spec.podCIDR`:

```sh
grep cluster-cidr /etc/kubernetes/manifests/kube-controller-manager.yaml
kubectl get nodes -o custom-columns=NAME:.metadata.name,PODCIDR:.spec.podCIDR
```

Don't confuse the Pod CIDR with the **Service** CIDR (`--service-cluster-ip-range` on the API
server, `10.96.0.0/16` here). ClusterIPs come from that one.
