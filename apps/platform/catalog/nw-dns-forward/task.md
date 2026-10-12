---
id: nw-dns-forward
title: Forward a domain in CoreDNS
domain: networking
host: cka-sim-control-plane
topics: [coredns, dns, configmaps]
weight: 6
order: last
---
## Context

The company's internal DNS zone `corp.internal` is served by Service `corp-dns` in Namespace
`nw-dns-forward` (port 53, UDP and TCP). Pods can't resolve names in that zone yet.

## Task

Change the cluster's CoreDNS configuration so that Pods resolve every name under
`corp.internal` through Service `corp-dns`. For example, `db.corp.internal` must resolve from Pod
`client` in Namespace `nw-dns-forward`.

All other DNS lookups, like `kubernetes.default.svc.cluster.local`, must keep working.
