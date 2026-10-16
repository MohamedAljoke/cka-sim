---
id: nw-gateway-migrate
title: Migrate an Ingress to Gateway API
domain: networking
host: cka-sim-control-plane
topics: [gateway-api, ingress, tls, migration]
weight: 7
---
## Context

Ingress `shop` in Namespace `nw-gateway-migrate` (IngressClass `nginx`) serves
`shop.nw-gateway-migrate.test` over HTTP and HTTPS with TLS Secret `shop-tls`:

- `/api` (Prefix) goes to Service `api`, port `80`
- `/` (Prefix) goes to Service `store`, port `80`

## Task

Move it to Gateway API, keeping the same behaviour, in Namespace `nw-gateway-migrate`:

1. Create a Gateway `shop-gw` with GatewayClass `traefik` and two listeners: `http` (HTTP, port
   `80`) and `https` (HTTPS, port `443`, hostname `shop.nw-gateway-migrate.test`, terminating TLS
   with Secret `shop-tls`).
2. Create an HTTPRoute `shop`, attached to `shop-gw`, with the same host and paths as the Ingress.
3. Delete Ingress `shop`.
