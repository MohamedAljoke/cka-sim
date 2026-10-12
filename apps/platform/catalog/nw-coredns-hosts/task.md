---
id: nw-coredns-hosts
title: Cluster-wide static DNS names
domain: networking
host: cka-sim-control-plane
topics: [coredns, dns, hosts-plugin]
weight: 5
order: last
---
## Context

Every app in the cluster must reach the old database server as `legacy-db.corp`, at `10.10.10.10`.
No DNS server knows that name.

## Task

Change the cluster's CoreDNS configuration so that `legacy-db.corp` resolves to `10.10.10.10` from
every Pod, for example from Pod `client` in Namespace `nw-coredns-hosts`. All other names must keep
resolving as before.
