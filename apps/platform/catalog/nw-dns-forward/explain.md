The Corefile is a list of **server blocks**. Each one starts with the zone it serves and the port.
CoreDNS answers a query from the block with the **most specific** zone that matches the name, so
adding a block for `corp.internal` sends those names elsewhere and leaves the main `.:53` block
alone:

```
corp.internal:53 {
    errors
    cache 30
    forward . 10.96.123.45        # corp-dns's ClusterIP; forward needs IPs, not names
}
```

```sh
kubectl -n nw-dns-forward get svc corp-dns                  # its ClusterIP
kubectl -n kube-system edit configmap coredns               # add the block after the .:53 block
kubectl -n kube-system logs -l k8s-app=kube-dns --tail=5    # "Reloading complete", or a parse error
kubectl -n nw-dns-forward exec client -- nslookup db.corp.internal
```

The `reload` plugin picks up the change within about 30 seconds. `kubectl -n kube-system rollout
restart deployment coredns` is quicker and shows a broken config right away.

The traps:
- `forward` takes IP addresses. CoreDNS can't resolve a name to find its own upstream.
- Don't put a second `forward` into the `.:53` block: one block can only have one.
- Always check that cluster names still resolve after a Corefile edit. A typo takes down DNS for
  the whole cluster.
