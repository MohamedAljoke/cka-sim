---
id: ar-subresources
title: Rights on subresources only
domain: architecture
host: cka-sim-control-plane
topics: [rbac, subresources, roles]
weight: 5
---
## Context

The support team, group `support`, works in Namespace `ar-subresources`. They need to read Pods,
read their logs, open a shell in them, and scale Deployments, but must not change Deployments in
any other way or delete Pods.

## Task

Create a Role `support` and a RoleBinding `support` in Namespace `ar-subresources` that let group
`support`:

- `get` and `list` Pods
- read Pod logs
- run `kubectl exec` in Pods
- `get`, `update` and `patch` the scale of Deployments

and nothing more.
