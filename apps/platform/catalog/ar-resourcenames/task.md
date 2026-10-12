---
id: ar-resourcenames
title: Access to one named object only
domain: architecture
host: cka-sim-control-plane
topics: [rbac, resourcenames, roles]
weight: 4
---
## Context

The feature-flag service in Namespace `ar-resourcenames` runs as ServiceAccount `toggler`. It must
read and change ConfigMap `feature-flags`, and must not touch any other ConfigMap, like `db-config`.

## Task

Create a Role `flags-editor` and a RoleBinding `toggler-flags` so that `toggler` can **get** and
**update** ConfigMap `feature-flags` and nothing else.
