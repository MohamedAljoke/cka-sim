---
id: nw-gateway-refgrant
title: Route to a Service in another namespace
domain: networking
host: cka-sim-control-plane
topics: [gateway-api, referencegrant, namespaces]
weight: 4
---
## Context

HTTPRoute `catalog` in Namespace `nw-gateway-refgrant` (host `catalog.nw-gateway-refgrant.test`,
Gateway `gw`) sends traffic to Service `catalog` in Namespace `nw-gateway-refgrant-data`. Requests
fail, because the data team hasn't allowed that reference.

## Task

Create a ReferenceGrant `allow-web-routes` in Namespace `nw-gateway-refgrant-data` that lets
HTTPRoutes from Namespace `nw-gateway-refgrant` (and only from there) reference Services in
`nw-gateway-refgrant-data`.
