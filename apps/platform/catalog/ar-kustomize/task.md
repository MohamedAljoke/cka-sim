---
id: ar-kustomize
title: Kustomize overlay for production
domain: architecture
host: cka-sim-control-plane
topics: [kustomize, overlays, deployments]
weight: 5
---
## Context

`/opt/course/ar-kustomize/base/` holds a Kustomize base for Deployment `shop` (1 replica, image
`nginx:1.26-alpine`).

## Task

Create a production overlay in `/opt/course/ar-kustomize/overlays/prod/` that uses the base
unchanged and:

- puts everything in Namespace `ar-kustomize`
- runs `3` replicas
- uses image `nginx:1.27-alpine`
- adds the label `env: prod` to the Deployment and its Pods

Apply the overlay with `kubectl apply -k`. Do not change the files in `base/`.
