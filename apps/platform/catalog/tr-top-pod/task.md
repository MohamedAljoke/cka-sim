---
id: tr-top-pod
title: Find the hungriest Pods
domain: troubleshooting
host: cka-sim-control-plane
topics: [metrics-server, kubectl-top]
weight: 3
---
## Context

The nodes are busy and the team suspects Namespace `tr-top-pod`. The cluster runs metrics-server.

## Task

1. Write the name of the Pod in Namespace `tr-top-pod` that uses the most **CPU** to
   `/opt/course/tr-top-pod/cpu.txt`.
2. Write the name of the Pod in Namespace `tr-top-pod` that uses the most **memory** to
   `/opt/course/tr-top-pod/memory.txt`.
