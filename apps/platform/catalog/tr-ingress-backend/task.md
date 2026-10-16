---
id: tr-ingress-backend
title: The shop's Ingress returns errors
domain: troubleshooting
host: cka-sim-control-plane
topics: [ingress, services, debugging]
weight: 5
---
## Context

Ingress `store` in Namespace `tr-ingress-backend` (IngressClass `nginx`, controller at
`10.96.80.80`) serves `store.tr-ingress-backend.test`:

- `/` should reach Service `storefront`
- `/api` should reach Service `api`

Both paths return errors instead. The Services and their Pods are healthy.

## Task

Find what's wrong in Ingress `store` and fix it so that both paths reach their Services. Don't
change, rename or add Services.
