---
id: wl-subpath
title: ConfigMap updates never arrive
domain: workloads
host: cka-sim-control-plane
topics: [configmaps, volumes, subpath]
weight: 4
---
## Context

Deployment `app` in Namespace `wl-subpath` reads `/etc/app/settings.conf` from ConfigMap
`app-config`. The ConfigMap was changed to `level=debug`, but the running app still reads
`level=info`, even after waiting.

## Task

Change Deployment `app` so that `/etc/app/settings.conf` follows changes to ConfigMap `app-config`
in running Pods, without a restart. The file path must stay the same.
