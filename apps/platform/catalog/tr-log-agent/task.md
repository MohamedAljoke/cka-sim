---
id: tr-log-agent
title: Node log agent as a DaemonSet
domain: troubleshooting
host: cka-sim-control-plane
topics: [logs, daemonsets, hostpath]
weight: 5
---
## Context

Pod `app` in Namespace `tr-log-agent-apps` logs to stdout. The platform team wants a node-level
log agent instead of reading logs Pod by Pod.

## Task

In Namespace `tr-log-agent`, create a DaemonSet `log-agent` (image `busybox:1.36`, label
`app=log-agent`) that runs on every **worker** node and:

- mounts the node's `/var/log/pods` **read-only** at `/var/log/pods`
- prints every log line of the Pods in Namespace `tr-log-agent-apps` on its node to its own stdout,
  for example with `tail -F /var/log/pods/tr-log-agent-apps_*/*/*.log`

`kubectl -n tr-log-agent logs` of the agent Pod on `app`'s node must show `app`'s log lines.
