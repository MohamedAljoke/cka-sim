---
id: nw-hostaliases
title: Static host entries in a Pod
domain: networking
host: cka-sim-control-plane
topics: [hostaliases, dns, pods]
weight: 2
---
## Context

A legacy app needs the names `db.legacy.local` and `cache.legacy.local`, which no DNS server
knows. Both run at `10.10.0.5`.

## Task

Create Pod `legacy` (image `busybox:1.36`, command `sleep 1d`) in Namespace `nw-hostaliases` in
which both names resolve to `10.10.0.5` through the Pod's `/etc/hosts`, without changing cluster
DNS.
