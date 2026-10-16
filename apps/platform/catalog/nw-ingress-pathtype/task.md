---
id: nw-ingress-pathtype
title: The docs pages return 404
domain: networking
host: cka-sim-control-plane
topics: [ingress, pathtype, debugging]
weight: 3
---
## Context

Ingress `docs` in Namespace `nw-ingress-pathtype` (IngressClass `nginx`, controller at
`10.96.80.80`) should send `http://docs.nw-ingress-pathtype.test/docs` and every page under it,
such as `/docs/guide`, to Service `docs`. `/docs` works, but `/docs/guide` returns 404.

## Task

Fix Ingress `docs` so that `/docs` and every path under `/docs/` reach Service `docs`, while
`/documents` does not.
