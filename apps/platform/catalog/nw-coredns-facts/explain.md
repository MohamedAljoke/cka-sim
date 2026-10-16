Cluster DNS is CoreDNS, run as Deployment `coredns` behind a Service that kubeadm still names
`kube-dns` for compatibility. Its whole configuration is the Corefile in ConfigMap `coredns`,
mounted into the Pods:

```sh
kubectl -n kube-system get svc kube-dns                  # 10.96.0.10, the IP in every Pod's resolv.conf
kubectl -n kube-system get deploy coredns -o yaml | grep -A3 configMap
kubectl -n kube-system get cm coredns -o yaml
```

```
.:53 {
    kubernetes cluster.local in-addr.arpa ip6.arpa {...}   # answers cluster names
    forward . /etc/resolv.conf                              # everything else goes upstream
    cache 30
    reload                                                   # re-reads the Corefile when it changes
    ...
}
```

`forward . /etc/resolv.conf` means "ask whatever nameservers the node uses". CoreDNS runs with
`dnsPolicy: Default`, so its own resolv.conf comes from the node. Replace it with IPs
(`forward . 8.8.8.8 1.1.1.1`) to send external lookups somewhere specific.
