CoreDNS logs only errors by default. The **`log`** plugin adds one line per query: who asked, what
type, which name, and the answer code.

```
.:53 {
    errors
    log                 # add this line inside the server block
    ...
}
```

```sh
kubectl -n kube-system edit cm coredns
kubectl -n kube-system rollout restart deploy coredns          # or wait ~30s for the reload plugin
kubectl -n nw-coredns-log exec client -- nslookup inventory.nw-coredns-log.svc.cluster.local
kubectl -n kube-system logs -l k8s-app=kube-dns --tail=50 | grep inventory
#   [INFO] 10.244.1.9:48311 - 12345 "A IN inventory.nw-coredns-log.svc.cluster.local. udp 60 false 512" NXDOMAIN qr,aa,rd 153 0.0001s
```

The line shows the client Pod's IP and port, the record type, the full name, and the result
(`NOERROR`, `NXDOMAIN`, `SERVFAIL`). It proves whether a query reached CoreDNS at all, and what it
answered. With `ndots:5` you'll also see the search-suffix attempts.

Query logs are verbose: every lookup of every Pod. Turn the plugin off again when you're done.
`log` can also filter, e.g. `log . { class denial error }` logs only failed lookups.
