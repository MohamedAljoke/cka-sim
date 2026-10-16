---
id: tr-runtime
title: Node lost its container runtime
domain: troubleshooting
host: cka-sim-worker2
topics: [kubelet, containerd, cri, nodes]
weight: 6
order: last
---
## Context

Node `cka-sim-worker2` reports `NotReady`. Its container runtime is containerd.

## Task

Find out why and fix it, so that node `cka-sim-worker2` is `Ready` again. The fix must survive a
restart of the kubelet.
