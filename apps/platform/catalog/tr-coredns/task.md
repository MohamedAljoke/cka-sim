---
id: tr-coredns
title: DNS lookups fail in Pods
domain: troubleshooting
host: cka-sim-control-plane
topics: [coredns, dns, configmaps]
weight: 6
order: last
---
## Context

Since a change to the cluster DNS, Pods can't resolve any name. Pod `client` in Namespace
`tr-coredns` shows it: `nslookup kubernetes.default.svc.cluster.local` times out.

## Task

Fix cluster DNS, so that Pod `client` can resolve `kubernetes.default.svc.cluster.local` again and
both CoreDNS Pods are Ready.

Keep CoreDNS's existing configuration apart from the fix.
