A cluster has two independent address ranges, set in different places:

| Range | Used for | Set by |
|---|---|---|
| **Pod CIDR** | Pod IPs | kube-controller-manager `--cluster-cidr` (kubeadm `networking.podSubnet`), split into one `spec.podCIDR` per node |
| **Service CIDR** | ClusterIPs | kube-apiserver `--service-cluster-ip-range` (kubeadm `networking.serviceSubnet`) |

```sh
grep cluster-cidr /etc/kubernetes/manifests/kube-controller-manager.yaml
grep service-cluster-ip-range /etc/kubernetes/manifests/kube-apiserver.yaml
kubectl -n kube-system get cm kubeadm-config -o yaml | grep -E 'podSubnet|serviceSubnet'
kubectl get nodes -o custom-columns=NAME:.metadata.name,CIDR:.spec.podCIDR
kubectl get svc kubernetes -n default        # its ClusterIP is the first address of the Service CIDR
```

They must not overlap with each other or with the nodes' own network. Pod IPs are real addresses
that the CNI routes. ClusterIPs are virtual and only exist as kube-proxy rules. A CNI may bring its
own IP pools and ignore the per-node `podCIDR`; Calico does unless configured otherwise.
