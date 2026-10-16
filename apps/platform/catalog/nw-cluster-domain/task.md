---
id: nw-cluster-domain
title: Cluster domain and DNS server
domain: networking
host: cka-sim-control-plane
topics: [dns, kubelet, coredns]
weight: 2
---
## Task

Find out which DNS settings the kubelet hands to Pods in this cluster, and write them to
`/opt/course/nw-cluster-domain/dns.txt`:

```
clusterDomain: <the cluster domain>
nameserver: <the DNS server IP Pods use>
```
