---
id: wl-host-process
title: A container is just a process
domain: workloads
host: cka-sim-control-plane
topics: [containers, crictl, security-context, processes]
weight: 4
---
## Context

Pod `proc` in Namespace `wl-host-process` runs on `cka-sim-control-plane` with `runAsUser: 3000`.

## Task

On the node, find the main process of Pod `proc`'s container and write these lines to
`/opt/course/wl-host-process/process.txt`:

```
pid: <its PID on the host>
uid: <the user ID it runs as, as the host sees it>
command: <its command line, as ps shows it>
```
