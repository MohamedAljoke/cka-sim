---
id: ar-crd-list
title: Find CRDs and their resources
domain: architecture
host: cka-sim-control-plane
topics: [crds, custom-resources, api-resources]
weight: 3
---
## Context

A monitoring operator installed several CustomResourceDefinitions in the API group
`monitoring.cka.io`.

## Task

1. Write the names of all CRDs of the group `monitoring.cka.io`, one per line, to
   `/opt/course/ar-crd-list/crds.txt`.
2. Write every `Dashboard` custom resource in the cluster, one per line as `<namespace>/<name>`,
   to `/opt/course/ar-crd-list/dashboards.txt`.
