---
id: tr-log-rotation
title: Rotate container logs on a node
domain: troubleshooting
host: cka-sim-worker2
topics: [kubelet, logs, log-rotation]
weight: 4
---
## Context

Chatty containers fill the disk of node `cka-sim-worker2` with their logs under `/var/log/pods/`.

## Task

Configure the kubelet of `cka-sim-worker2` in its config file so that each container's log file
is rotated at `5Mi` and at most `3` log files are kept per container. Apply the change so the
running kubelet uses it, and make sure the node is `Ready`.
