---
id: ar-rbac
title: ServiceAccount with least privilege
domain: architecture
host: cka-sim-control-plane
topics: [rbac, serviceaccounts]
weight: 6
---
## Context

A CI pipeline will deploy into Namespace `ci` using its own identity.

## Task

1. Create a ServiceAccount named `deployer` in Namespace `ci`.
2. Create a Role named `deployer` in Namespace `ci` that allows only `get`, `list`, `create`
   and `update` on `deployments`.
3. Bind the Role to the ServiceAccount with a RoleBinding named `deployer`.
4. Check whether the ServiceAccount may **delete** Deployments in `ci` using `kubectl auth can-i`,
   and write its output (`yes` or `no`) to `/opt/course/ar-rbac/can-delete.txt`.
