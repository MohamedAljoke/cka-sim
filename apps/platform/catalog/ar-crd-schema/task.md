---
id: ar-crd-schema
title: Write a CRD with validation
domain: architecture
host: cka-sim-control-plane
topics: [crds, openapi-schema, validation]
weight: 6
---
## Task

1. Write and install a CustomResourceDefinition for kind `Database` in group `cka.example.com`,
   version `v1`, **namespaced**, plural `databases`, short name `db`, with this schema for `spec`:
   - `engine`: a **required** string, either `postgres` or `mysql`
   - `storageGB`: an integer, at least `1`

   `kubectl get databases` must show two extra columns, `Engine` (`.spec.engine`) and `Storage`
   (`.spec.storageGB`).
2. In Namespace `ar-crd-schema`, create Database `orders` with engine `postgres` and storageGB `10`.
3. Try to create Database `legacy` with engine `oracle`. It must be rejected. Write the error
   message to `/opt/course/ar-crd-schema/rejected.txt`.
