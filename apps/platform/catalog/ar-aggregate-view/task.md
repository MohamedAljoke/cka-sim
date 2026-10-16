---
id: ar-aggregate-view
title: Extend the built-in view role
domain: architecture
host: cka-sim-control-plane
topics: [rbac, aggregation, clusterroles, crds]
weight: 5
---
## Context

The cluster has a new CRD `widgets.cka.example.com`. User `reader` has the built-in ClusterRole
`view` in Namespace `ar-aggregate`, but `kubectl get widgets` is forbidden for them. Everyone with
`view` anywhere should be able to read widgets.

## Task

Create a ClusterRole `widget-viewer` that grants `get`, `list` and `watch` on widgets, and make the
built-in `view` ClusterRole include it **automatically**, without editing `view` itself.
