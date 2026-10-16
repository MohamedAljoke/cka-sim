---
id: nw-netpol-ipblock
title: Allow a CIDR except one address
domain: networking
host: cka-sim-control-plane
topics: [networkpolicy, ipblock]
weight: 4
---
## Context

Pod `web` in Namespace `nw-netpol-ipblock` (label `app=web`) listens on TCP `80`. Pods `near`, `banned`
and `far` in the same Namespace call it.

`/opt/course/nw-netpol-ipblock/cidr` holds a Pod CIDR, and `/opt/course/nw-netpol-ipblock/except` holds
one IP address inside it.

## Task

Create a NetworkPolicy `web-ipblock` in Namespace `nw-netpol-ipblock` so that `web` accepts TCP `80`
only from addresses in that CIDR, **except** that one IP address. Use `ipBlock`.
