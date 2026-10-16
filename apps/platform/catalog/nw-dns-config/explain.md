With the default `dnsPolicy: ClusterFirst`, the kubelet writes each Pod's `/etc/resolv.conf` like
this:

```
search nw-dns-config.svc.cluster.local svc.cluster.local cluster.local
nameserver 10.96.0.10
options ndots:5
```

- The **search** list is why `api` (same namespace) and `api.other` (another namespace) resolve:
  the resolver tries each suffix in turn.
- **`ndots:5`** means any name with fewer than 5 dots is tried with every search suffix *first*. A
  lookup of `www.example.com` (2 dots) costs 3 failed cluster queries before the real one. Lower
  `ndots`, or a trailing dot (`www.example.com.`), avoids that for apps that mostly call external
  names.

`dnsPolicy: None` ignores all of that and uses only `dnsConfig`. `dnsConfig` can also be combined
with `ClusterFirst` to add or override single options:

```yaml
dnsPolicy: None
dnsConfig:
  nameservers: [10.96.0.10]
  searches: [nw-dns-config.svc.cluster.local]
  options:
  - {name: ndots, value: "2"}
```

The other policies: `Default` takes the node's own resolv.conf (no cluster names at all), and
`ClusterFirstWithHostNet` is ClusterFirst for Pods with `hostNetwork: true`.
