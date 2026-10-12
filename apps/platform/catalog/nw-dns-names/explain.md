CoreDNS answers two kinds of names under the cluster domain (`cluster.local` unless someone
changed it):

| Name | Resolves to |
|---|---|
| `<service>.<namespace>.svc.cluster.local` | the Service's ClusterIP (or the Pod IPs for a headless Service) |
| `<pod-ip-with-dashes>.<namespace>.pod.cluster.local` | that Pod IP |

```sh
kubectl -n nw-dns-names exec worker -- nslookup api.nw-dns-names.svc.cluster.local
kubectl -n nw-dns-names get pod worker -o wide         # 10.244.2.14 -> 10-244-2-14
kubectl -n nw-dns-names exec worker -- nslookup 10-244-2-14.nw-dns-names.pod.cluster.local
kubectl -n nw-dns-names exec worker -- cat /etc/resolv.conf   # search nw-dns-names.svc.cluster.local svc.cluster.local ...
```

Thanks to the `search` list in each Pod's `/etc/resolv.conf`, `api` works in the same namespace
and `api.nw-dns-names` works from anywhere. The full name always works, which is why exam answers
should use it.

Pod names by IP only exist because the Corefile's `kubernetes` plugin has `pods insecure`. Pods
of a StatefulSet behind a headless Service also get stable names:
`<pod>.<service>.<namespace>.svc.cluster.local`.
