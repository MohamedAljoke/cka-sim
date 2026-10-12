---
id: nw-headless
title: Headless Service DNS
domain: networking
host: cka-sim-control-plane
topics: [services, headless, dns]
weight: 3
---
## Context

Deployment `db` in Namespace `nw-headless` runs 3 Pods labelled `app=db`. The client library wants
to connect to every Pod itself, not through one virtual IP.

## Task

1. Create a **headless** Service `db` that selects `app=db` on port `80`.
2. From Pod `client`, look up `db.nw-headless.svc.cluster.local` and write the IP addresses it
   returns, one per line, to `/opt/course/nw-headless/ips.txt`.
