---
id: ar-helm-template
title: Use a chart without Helm
domain: architecture
host: cka-sim-control-plane
topics: [helm, template, manifests]
weight: 4
---
## Context

The platform team wants to use the chart `web` (version `1.0.0`, repository `http://cka-sim-charts`),
but without Helm managing anything in the cluster.

## Task

1. Render chart `web` version `1.0.0` with release name `docs` for Namespace `ar-helm-template`
   into the file `/opt/course/ar-helm-template/docs.yaml`.
2. Create the objects from that file in Namespace `ar-helm-template` with `kubectl`. There must be
   no Helm release in that Namespace.
