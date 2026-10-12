---
id: nw-coredns-facts
title: Where CoreDNS gets its config
domain: networking
host: cka-sim-control-plane
topics: [coredns, dns, configmaps]
weight: 2
---
## Task

Write these facts about the cluster DNS to `/opt/course/nw-coredns-facts/dns.txt`:

```
service-ip: <ClusterIP of the cluster DNS Service>
configmap: <namespace/name of the ConfigMap holding the Corefile>
upstream: <where CoreDNS forwards names it isn't authoritative for, as written in the Corefile>
```
