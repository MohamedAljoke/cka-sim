---
id: nw-netpol-lockdown
title: Default-deny a namespace but keep DNS
domain: networking
host: cka-sim-control-plane
topics: [networkpolicy, default-deny, egress, dns]
weight: 5
---
## Context

Namespace `nw-netpol-lockdown` runs Pod `app` and Pod `web` (listening on TCP `8080`). Namespace
`nw-netpol-visitor` runs Pod `visitor`. Nothing restricts traffic yet.

## Task

In Namespace `nw-netpol-lockdown`:

1. Create a NetworkPolicy `default-deny` that blocks **all** ingress and **all** egress for every
   Pod in the Namespace.
2. Create a NetworkPolicy `allow-dns` that lets every Pod in the Namespace send DNS queries
   (UDP and TCP port `53`), and nothing else.
