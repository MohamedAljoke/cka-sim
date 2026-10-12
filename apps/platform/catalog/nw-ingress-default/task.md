---
id: nw-ingress-default
title: Catch unmatched requests with a default backend
domain: networking
host: cka-sim-control-plane
topics: [ingress, defaultbackend]
weight: 3
---
## Context

Ingress `site` in Namespace `nw-ingress-default` (IngressClass `nginx`, controller at `10.96.80.80`)
sends host `www.nw-ingress-default.test` to Service `site`. Requests for any other host get a 404
from the controller. Service `fallback` in the same Namespace should answer those instead.

## Task

Change Ingress `site` so that requests matching none of its rules go to Service `fallback` on port
`80`. Keep the existing rule.
