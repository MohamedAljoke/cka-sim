---
id: tr-top-containers
title: Which container eats the memory
domain: troubleshooting
host: cka-sim-control-plane
topics: [metrics-server, kubectl-top, multi-container]
weight: 3
---
## Context

Pod `app` in Namespace `tr-top-containers` runs three containers and uses far more memory than
expected. The cluster runs metrics-server.

## Task

Write the name of the **container** in Pod `app` that uses the most memory to
`/opt/course/tr-top-containers/container.txt`.
