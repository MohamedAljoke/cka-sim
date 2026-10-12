---
id: nw-netpol-ports
title: Open a port range and a named port
domain: networking
host: cka-sim-control-plane
topics: [networkpolicy, endport, named-ports]
weight: 4
---
## Context

Pod `server` in Namespace `nw-netpol-ports` (label `app=server`) listens on TCP `8000`, `8005`,
`8010`, `8011` and `9100`. Its container declares port `9100` with the name `metrics`. Pod `client`
in the same Namespace calls it.

## Task

Create a NetworkPolicy `server-ports` in Namespace `nw-netpol-ports` that lets any Pod reach `server`
on:

- TCP `8000` to `8010`, written as **one** port entry using `endPort`
- the container port named `metrics`, referenced **by its name**, not its number

Every other port on `server` must be blocked.
