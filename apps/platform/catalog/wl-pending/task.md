---
id: wl-pending
title: Fix Pods stuck in Pending
domain: workloads
host: cka-sim-control-plane
topics: [scheduling, troubleshooting, node-affinity, resources]
weight: 4
---
## Context

In Namespace `wl-pending`, the Pods of Deployments `api` and `batch` never start.

## Task

Find out why each Pod is `Pending` and fix each Deployment so it runs **1** Ready Pod.

- Don't add labels or taints to nodes.
- Keep the containers and their images as they are.
- `batch` must request at most `256Mi` of memory.
