---
id: nw-gateway-header
title: Route by request header
domain: networking
host: cka-sim-control-plane
topics: [gateway-api, httproute, header-matching]
weight: 4
---
## Context

HTTPRoute `app` in Namespace `nw-gateway-header` sends every request for
`app.nw-gateway-header.test` to Service `v1` (port `80`) through Gateway `gw`. Service `v2` runs
the next version.

## Task

Change HTTPRoute `app` so that requests with header `x-version: v2` go to Service `v2` on port
`80`. All other requests keep going to `v1`.
