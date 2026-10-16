---
id: nw-ingress-rules
title: Route two paths with an Ingress
domain: networking
host: cka-sim-control-plane
topics: [ingress, ingressclass]
weight: 4
---
## Context

Namespace `nw-ingress-rules` runs two Services on port `80`: `store` and `api`. The cluster's
ingress controller serves the IngressClass `nginx` and listens on `10.96.80.80`.

## Task

1. Create an Ingress `shop` in Namespace `nw-ingress-rules` with `ingressClassName: nginx` and host
   `shop.nw-ingress-rules.test`:
   - path `/api` (pathType `Prefix`) goes to Service `api`, port `80`
   - path `/` (pathType `Prefix`) goes to Service `store`, port `80`
2. Test it with `curl`, sending the host in a `Host` header. Save the response to
   `http://shop.nw-ingress-rules.test/api/orders` in `/opt/course/nw-ingress-rules/curl.txt`.
