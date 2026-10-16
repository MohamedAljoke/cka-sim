---
id: nw-ingress-cross
title: An Ingress can't reach another namespace
domain: networking
host: cka-sim-control-plane
topics: [ingress, namespaces, debugging]
weight: 4
---
## Context

Ingress `portal` in Namespace `nw-ingress-cross` (IngressClass `nginx`, controller at `10.96.80.80`)
serves host `portal.nw-ingress-cross.test`:

- `/` goes to Service `portal`, which works
- `/billing` should go to Service `billing`, but that Service lives in Namespace
  `nw-ingress-cross-billing`, and requests to `/billing` fail

## Task

Make `http://portal.nw-ingress-cross.test/billing` and the paths under it reach Service `billing`
in `nw-ingress-cross-billing`, keeping `/` on `portal`:

1. Create an Ingress `billing` in Namespace `nw-ingress-cross-billing` for the same host and path
   `/billing` (pathType `Prefix`).
2. Remove the broken `/billing` path from Ingress `portal`.

Don't move or copy the Services.
