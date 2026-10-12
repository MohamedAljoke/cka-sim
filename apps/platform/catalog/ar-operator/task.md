---
id: ar-operator
title: Install an operator and let it build the app
domain: architecture
host: cka-sim-control-plane
topics: [operators, crds, custom-resources]
weight: 5
---
## Context

`/opt/course/ar-operator/web-operator.yaml` installs **web-operator**: a CustomResourceDefinition
for `WebApp` (group `cka-sim.io`, version `v1`) and the operator itself. For each `WebApp`, the
operator creates and keeps a Deployment and a Service of the same name, built from the `WebApp`'s
`spec.image`, `spec.replicas` and `spec.port`.

## Task

1. Install the operator from that file, and wait until its Deployment in Namespace `web-operator`
   is available.
2. Create a `WebApp` named `shop` in Namespace `ar-operator` with image `nginx:1.27-alpine`, 2
   replicas and port `80`.
3. Don't create the Deployment or Service yourself: the operator must build them.
