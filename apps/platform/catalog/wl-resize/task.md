---
id: wl-resize
title: Resize a running Pod in place
domain: workloads
host: cka-sim-control-plane
topics: [in-place-resize, resources, pods]
weight: 4
---
## Context

Pod `worker` in Namespace `wl-resize` needs more CPU, but restarting it would lose an hour of
in-memory work.

## Task

Change Pod `worker`'s container `worker` to a CPU request of `300m` and a CPU limit of `400m`
**without restarting or recreating** the Pod or its container.
