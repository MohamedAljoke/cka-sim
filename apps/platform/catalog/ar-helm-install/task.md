---
id: ar-helm-install
title: Install, upgrade and roll back a chart
domain: architecture
host: cka-sim-control-plane
topics: [helm, charts, releases]
weight: 6
---
## Context

The team's Helm charts are served from the repository `http://cka-sim-charts`. It has chart `web`
in versions `1.0.0` and `1.1.0`.

## Task

1. Add the repository as `cka`.
2. Install chart `cka/web` version `1.0.0` as release `shop` in Namespace `ar-helm-install`, with
   the value `replicaCount` set to `3`.
3. Upgrade release `shop` to chart version `1.1.0`, keeping `replicaCount` at `3`.
4. Then roll release `shop` back to its first revision.
