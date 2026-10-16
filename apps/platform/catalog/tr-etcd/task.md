---
id: tr-etcd
title: Control plane down, find the logs
domain: troubleshooting
host: cka-sim-control-plane
topics: [etcd, crictl, logs, static-pods]
weight: 7
order: last
---
## Context

`kubectl` on `cka-sim-control-plane` stopped working. One of the control-plane components keeps
crashing, and it isn't the API server's own configuration that is wrong.

## Task

1. Using the container runtime on the node, find the component that fails and write the log line
   in which it reports its fatal error to `/opt/course/tr-etcd/error.log`.
2. Fix it, so that `kubectl get nodes` works again.
