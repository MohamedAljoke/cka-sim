---
id: nw-netpol-ingress
title: Let only the frontend and monitoring reach the API
domain: networking
host: cka-sim-control-plane
topics: [networkpolicy, namespaceselector, podselector]
weight: 5
---
## Context

Namespace `nw-netpol-ingress` runs Pod `api` (label `app=api`), listening on TCP `8080` and `9090`,
and two clients: `frontend` (label `role=frontend`) and `worker` (label `role=worker`). Namespace
`nw-netpol-monitor` runs Pod `probe`, and Namespace `nw-netpol-outside` runs Pod `stranger`. Right
now every Pod can reach `api` on both ports.

## Task

Create a NetworkPolicy `api-allow` in Namespace `nw-netpol-ingress` so that `api` accepts traffic
only on TCP `8080`, and only from:

- Pods labelled `role=frontend` in Namespace `nw-netpol-ingress`
- any Pod in Namespace `nw-netpol-monitor`

Everything else that reaches `api` must be blocked.
