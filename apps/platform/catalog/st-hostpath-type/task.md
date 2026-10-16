---
id: st-hostpath-type
title: hostPath volume won't mount
domain: storage
host: cka-sim-control-plane
topics: [hostpath, volumes, troubleshooting]
weight: 3
---
## Context

Pod `agent` in Namespace `st-hostpath-type` (created from `/opt/course/st-hostpath-type/agent.yaml`)
is stuck in `ContainerCreating`. It runs on `cka-sim-control-plane` and keeps its state in the
node's `/var/lib/agent-data`.

## Task

Fix `agent.yaml` so that the kubelet **creates** `/var/lib/agent-data` when it's missing, and
recreate Pod `agent` from it. Keep the hostPath path. Don't create the directory by hand.
