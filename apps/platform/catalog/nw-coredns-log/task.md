---
id: nw-coredns-log
title: Log DNS queries in CoreDNS
domain: networking
host: cka-sim-control-plane
topics: [coredns, dns, logs, debugging]
weight: 5
order: last
---
## Context

A team claims their app never sends DNS queries for `inventory.nw-coredns-log.svc.cluster.local`.
Pod `client` in Namespace `nw-coredns-log` runs that app.

## Task

1. Turn on query logging in CoreDNS.
2. Run `nslookup inventory.nw-coredns-log.svc.cluster.local` in Pod `client`, find the CoreDNS log
   line for that query, and write it to `/opt/course/nw-coredns-log/query.log`.
3. Turn query logging off again. Leave the rest of the Corefile as it was.
