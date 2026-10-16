The kubelet writes every Pod's `/etc/resolv.conf` from two settings in its config file:
`clusterDNS`, the IP of the `kube-dns` Service, and `clusterDomain`, which builds the `search`
list. CoreDNS's `kubernetes` plugin has to serve the same domain, or names won't resolve.

```sh
grep -A1 -E 'clusterDomain|clusterDNS' /var/lib/kubelet/config.yaml
kubectl -n kube-system get svc kube-dns                      # same IP as clusterDNS
kubectl run tmp --image=busybox:1.36 --restart=Never --rm -i -- cat /etc/resolv.conf
#   search default.svc.cluster.local svc.cluster.local cluster.local
#   nameserver 10.96.0.10
#   options ndots:5
kubectl -n kube-system get cm coredns -o yaml | grep kubernetes   # kubernetes cluster.local ...
```

Changing the cluster domain means changing it in all three places: every kubelet's config, the
Corefile, and the API server's idea of it in the kubeadm ClusterConfiguration
(`networking.dnsDomain`). Then restart the kubelets and recreate the Pods. Existing Pods keep the
`resolv.conf` they started with.
