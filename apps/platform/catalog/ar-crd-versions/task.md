---
id: ar-crd-versions
title: Serve a CRD in two versions
domain: architecture
host: cka-sim-control-plane
topics: [crds, versions, storage-version]
weight: 5
---
## Context

CRD `gadgets.cka.example.com` serves only version `v1alpha1`, which is also its storage version.
The API is now stable.

## Task

1. Add version `v1` to the CRD with the same schema. Make `v1` the **storage** version, and keep
   serving `v1alpha1` (no longer stored).
2. Create Gadget `beacon` in Namespace `ar-crd-versions` through version `v1` (`spec.color: blue`).
3. Write the CRD's stored versions, as reported in its status, to
   `/opt/course/ar-crd-versions/stored.txt`, comma-separated.
