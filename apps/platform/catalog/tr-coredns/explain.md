Every Pod's `/etc/resolv.conf` points at the `kube-dns` Service (`10.96.0.10`), which selects the
CoreDNS Pods. CoreDNS reads its configuration, the **Corefile**, from ConfigMap `coredns` in
`kube-system`. Each line in a server block is a plugin; `kubernetes` is the one that answers for
`cluster.local` names. A misspelled plugin name is a parse error, and CoreDNS refuses to start.

```sh
kubectl -n tr-coredns exec client -- nslookup kubernetes.default.svc.cluster.local
kubectl -n kube-system get pods -l k8s-app=kube-dns               # CrashLoopBackOff
kubectl -n kube-system logs -l k8s-app=kube-dns --tail=5          # Unknown directive 'kubernetess'
kubectl -n kube-system edit configmap coredns
kubectl -n kube-system rollout restart deployment coredns
```

The trap is the `reload` plugin: a running CoreDNS re-reads a changed Corefile on its own, but a
broken config only shows up when Pods restart. That's why DNS can work for a while after a bad
edit, and why you check `logs` after every Corefile change. After the fix, `rollout restart`
gives you fresh Pods right away instead of waiting for the crash-loop back-off.

Also check the path in between when DNS fails: `kubectl -n kube-system get endpointslices -l
kubernetes.io/service-name=kube-dns` (no endpoints means no Ready CoreDNS Pods), and the
`nameserver` in the Pod's `/etc/resolv.conf`.
