---
id: nw-netpol-and-or
title: The policy lets in too much
domain: networking
host: cka-sim-control-plane
topics: [networkpolicy, namespaceselector, podselector, debugging]
weight: 4
---
## Context

Pod `api` in Namespace `nw-netpol-and-or` listens on TCP `8080`. NetworkPolicy `api-from-monitoring`
was meant to let in only Pods labelled `app=prometheus` that run in a Namespace labelled
`team=monitoring`. Namespace `nw-netpol-mon` has that label and runs Pods `prometheus` and
`grafana`.

Yet `grafana` reaches `api`, and so does Pod `prometheus` in `nw-netpol-and-or` itself.

## Task

Fix NetworkPolicy `api-from-monitoring`, keeping its name, so that only `app=prometheus` Pods in
Namespaces labelled `team=monitoring` can reach `api`.
