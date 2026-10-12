CoreDNS's **`hosts`** plugin serves names from an `/etc/hosts`-style list, inline or from a file.
Put it in the main server block before `forward`:

```
.:53 {
    errors
    health { ... }
    ready
    hosts {
       10.10.10.10 legacy-db.corp
       fallthrough                # names not in the list go on to the next plugins
    }
    kubernetes cluster.local in-addr.arpa ip6.arpa { ... }
    forward . /etc/resolv.conf
    ...
}
```

```sh
kubectl -n kube-system edit cm coredns
kubectl -n kube-system rollout restart deploy coredns
kubectl -n nw-coredns-hosts exec client -- nslookup legacy-db.corp
```

The trap is forgetting `fallthrough`: without it, `hosts` answers NXDOMAIN for every other name in
the zones it covers. Here that's the root zone `.`, so all of cluster DNS breaks. The order of
plugins in the Corefile doesn't matter (CoreDNS has a fixed internal order), but readability does.

Alternatives: `hostAliases` (one Pod at a time), `rewrite name legacy-db.corp db.legacy.svc.cluster.local`
to alias to a Service, or an ExternalName / selectorless Service when an in-cluster name is fine.
