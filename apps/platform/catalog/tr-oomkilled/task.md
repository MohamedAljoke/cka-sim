---
id: tr-oomkilled
title: Container keeps getting killed
domain: troubleshooting
host: cka-sim-control-plane
topics: [oomkilled, resources, limits]
weight: 4
---
## Context

The Pod of Deployment `cache` in Namespace `tr-oomkilled` keeps restarting. Its logs show nothing
unusual. The app needs about `60Mi` of memory while it loads its data.

## Task

1. Write the reason the container was last terminated, as Kubernetes reports it, to
   `/opt/course/tr-oomkilled/reason.txt`.
2. Fix Deployment `cache` so that it runs stably: set its memory limit to `128Mi`. Leave its
   command and memory request unchanged.
