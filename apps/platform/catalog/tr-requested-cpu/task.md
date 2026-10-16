---
id: tr-requested-cpu
title: Which node has the most CPU requested
domain: troubleshooting
host: cka-sim-control-plane
topics: [resources, requests, nodes, scheduling]
weight: 3
---
## Context

There is no metrics-server in this cluster, but the team wants to know how full the nodes are from
the scheduler's point of view.

## Task

Write the name of the node with the highest total CPU **requests** of its Pods to
`/opt/course/tr-requested-cpu/node.txt`.
