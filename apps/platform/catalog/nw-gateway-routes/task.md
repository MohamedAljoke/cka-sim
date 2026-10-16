---
id: nw-gateway-routes
title: Route traffic with a Gateway and an HTTPRoute
domain: networking
host: cka-sim-control-plane
topics: [gateway-api, httproute, gatewayclass]
weight: 5
---
## Context

Namespace `nw-gateway-routes` runs Services `web` and `api` on port `80`. The cluster has the
Gateway API installed, with a GatewayClass `traefik` whose controller listens on `10.96.80.80`.

## Task

In Namespace `nw-gateway-routes`:

1. Create a Gateway `web-gw` with GatewayClass `traefik` and one listener named `http`: protocol
   `HTTP`, port `80`.
2. Create an HTTPRoute `web-route` attached to `web-gw` for hostname `app.nw-gateway-routes.test`:
   - paths starting with `/api` go to Service `api`, port `80`
   - everything else goes to Service `web`, port `80`
