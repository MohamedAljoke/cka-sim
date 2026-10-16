---
id: tr-cordon
title: Pods stuck on a disabled node
domain: troubleshooting
host: cka-sim-control-plane
topics: [nodes, cordon, scheduling]
weight: 3
order: last
---
## Context

Deployment `agent` in Namespace `tr-cordon` must run one Pod on each worker node, but one of its
Pods stays `Pending`.

## Task

1. Find out why, and get both Pods of `agent` scheduled without restarting anything.
2. Write the name of the node that caused the problem to `/opt/course/tr-cordon/node`.

Do not modify the Deployment `agent`.
