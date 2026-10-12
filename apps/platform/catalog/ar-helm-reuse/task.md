---
id: ar-helm-reuse
title: An upgrade lost the custom values
domain: architecture
host: cka-sim-control-plane
topics: [helm, values, upgrades]
weight: 5
---
## Context

Release `portal` in Namespace `ar-helm-reuse` ran 3 replicas behind a `NodePort` Service. A
colleague then ran:

```sh
helm -n ar-helm-reuse upgrade portal web --repo http://cka-sim-charts --version 1.1.0 --set image.pullPolicy=Always
```

Since then `portal` runs 2 replicas and its Service is `ClusterIP`.

## Task

Upgrade release `portal` in Namespace `ar-helm-reuse` so that it again has `replicaCount` `3` and
`service.type` `NodePort`, and still has `image.pullPolicy` `Always`. Stay on chart version
`1.1.0`, and use an upgrade, not a rollback.
