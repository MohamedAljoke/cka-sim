---
id: tr-top-nodes
title: Node usage, sorted by memory
domain: troubleshooting
host: cka-sim-control-plane
topics: [metrics-server, kubectl-top, nodes]
weight: 2
---
## Context

The on-call team wants a quick way to see which nodes are low on memory. The cluster runs
metrics-server.

## Task

Write a command to `/opt/course/tr-top-nodes/nodes.sh` that shows the CPU and memory usage of
every node, sorted by memory usage, highest first. Running `bash /opt/course/tr-top-nodes/nodes.sh`
must print the list in the `kubectl top` format.
