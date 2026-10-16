---
id: tr-debug
title: Debug a container with no shell
domain: troubleshooting
host: cka-sim-control-plane
topics: [kubectl-debug, ephemeral-containers, distroless]
weight: 4
---
## Context

Pod `app` in Namespace `tr-debug` runs a minimal image with no shell and no tools, so
`kubectl exec` can't help.

## Task

1. Attach an ephemeral debug container named `debugger` (image `busybox:1.36`) to Pod `app` that
   shares the process namespace of container `app`.
2. From it, list the processes of container `app` and write the output to
   `/opt/course/tr-debug/ps.txt`.

Do not restart or recreate Pod `app`.
