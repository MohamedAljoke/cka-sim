---
id: nw-ingress-rewrite
title: Strip a path prefix with rewrite-target
domain: networking
host: cka-sim-control-plane
topics: [ingress, annotations, rewrite]
weight: 4
---
## Context

Ingress `app` in Namespace `nw-ingress-rewrite` (IngressClass `nginx`) sends
`http://app.nw-ingress-rewrite.test/app` to Service `web`. But `web` expects requests at `/`, and
right now it receives `/app`. Its answer shows the path it got.

## Task

Change Ingress `app` so that a request to `/app` reaches `web` as `/`. Use the controller's
`nginx.ingress.kubernetes.io/rewrite-target` annotation.
