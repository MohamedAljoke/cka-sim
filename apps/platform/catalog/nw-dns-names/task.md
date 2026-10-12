---
id: nw-dns-names
title: DNS names of Services and Pods
domain: networking
host: cka-sim-control-plane
topics: [dns, coredns, services, pods]
weight: 3
---
## Context

Namespace `nw-dns-names` has Service `api` and Pod `worker`.

## Task

1. Write the fully qualified DNS name of Service `api` to `/opt/course/nw-dns-names/service.txt`.
2. Write the DNS name that resolves to Pod `worker`'s IP (the name built from its IP address) to
   `/opt/course/nw-dns-names/pod.txt`.

Check that both resolve with `nslookup` from a Pod.
