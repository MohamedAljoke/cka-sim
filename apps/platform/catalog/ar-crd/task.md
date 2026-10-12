---
id: ar-crd
title: Install a CRD and create a resource
domain: architecture
host: cka-sim-control-plane
topics: [crds, custom-resources, kubectl-explain]
weight: 4
---
## Task

1. Install the CustomResourceDefinition in `/opt/course/ar-crd/crd.yaml`.
2. In Namespace `ar-crd`, create a resource of its kind named `nightly`, with schedule
   `0 2 * * *` and a retention of `7`.
3. Write the field names that `kubectl explain` shows under `spec` of that kind, one per line, to
   `/opt/course/ar-crd/fields.txt`.
