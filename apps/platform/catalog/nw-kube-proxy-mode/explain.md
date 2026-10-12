In **iptables** mode, kube-proxy turns every Service into NAT rules, in a fixed chain structure
in the `nat` table:

```
KUBE-SERVICES                  match the ClusterIP:port      -> jump KUBE-SVC-<hash>    (one per Service port)
KUBE-SVC-<hash>                pick one endpoint at random   -> jump KUBE-SEP-<hash>    (one per endpoint)
KUBE-SEP-<hash>                DNAT to podIP:targetPort
```

Each rule carries a comment `namespace/service:port`, which is how you find them:

```sh
kubectl -n kube-system get cm kube-proxy -o yaml | grep mode     # iptables, ipvs or nftables ("" = iptables)
iptables-save -t nat | grep 'nw-kube-proxy-mode/api'
iptables-save -t nat | grep '^-A KUBE-SVC-XXXX'                    # --probability 0.5 per endpoint
iptables-save -t nat | grep '^-A KUBE-SEP-YYYY'                    # -j DNAT --to-destination 10.244.1.5:80
```

The SVC chain splits traffic with `statistic --mode random --probability`: 1/2 for the first of 2
endpoints, then everything left for the last. In **ipvs** mode, `ipvsadm -Ln` shows a virtual
server per ClusterIP instead. In **nftables** mode, look with `nft list table ip kube-proxy`.

If a Service has endpoints but no KUBE-SVC chain on a node, kube-proxy on that node isn't doing
its job.
