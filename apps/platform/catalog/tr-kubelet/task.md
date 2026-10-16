---
id: tr-kubelet
title: Worker node NotReady
domain: troubleshooting
host: cka-sim-worker
topics: [kubelet, systemd, nodes]
weight: 7
order: last
---
## Context

Node `cka-sim-worker` reports `NotReady` and no new Pods start on it.

## Task

Find out why and fix it, so that node `cka-sim-worker` is `Ready`.

The fix must survive a reboot of the node.
