---
id: wl-immutable
title: Immutable ConfigMap, new version
domain: workloads
host: cka-sim-control-plane
topics: [configmaps, immutable, rollouts]
weight: 4
---
## Context

Deployment `app` in Namespace `wl-immutable` reads `COLOR` from ConfigMap `app-config-v1`
(`color=green`).

## Task

1. Make ConfigMap `app-config-v1` immutable, with its data unchanged.
2. The app must now show `COLOR=blue`. Since `app-config-v1` can't change, create an immutable
   ConfigMap `app-config-v2` with `color=blue`, and roll Deployment `app` over to it.
