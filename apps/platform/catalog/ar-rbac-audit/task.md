---
id: ar-rbac-audit
title: Audit a user's bindings
domain: architecture
host: cka-sim-control-plane
topics: [rbac, bindings, least-privilege, audit]
weight: 5
---
## Context

User `contractor` should only be able to **read** resources in Namespace `ar-audit-app`. A review
suspects they have collected far more rights over time.

## Task

1. Find every RoleBinding and ClusterRoleBinding in the cluster that grants anything to user
   `contractor`, and write their names to `/opt/course/ar-rbac-audit/bindings.txt`, one per line.
2. Remove the access that goes beyond reading in `ar-audit-app`, so that `contractor` keeps
   exactly the `view` rights there and nothing else.
