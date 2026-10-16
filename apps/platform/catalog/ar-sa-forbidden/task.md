---
id: ar-sa-forbidden
title: ServiceAccount gets forbidden
domain: architecture
host: cka-sim-control-plane
topics: [rbac, serviceaccounts, roles]
weight: 4
---
## Context

The app in Deployment `reporter` (Namespace `ar-sa`) runs as ServiceAccount `reporter` and logs:

```
pods is forbidden: User "system:serviceaccount:ar-sa:reporter" cannot list resource "pods" in API group "" in the namespace "ar-sa"
```

## Task

Fix the RBAC setup so that ServiceAccount `reporter` can list Pods in Namespace `ar-sa`.

Grant nothing more than that: `reporter` must not be able to delete Pods, read Secrets, or list
Pods in other namespaces. Do not modify Deployment `reporter`.
