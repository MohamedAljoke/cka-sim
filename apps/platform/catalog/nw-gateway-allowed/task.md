---
id: nw-gateway-allowed
title: The route doesn't attach to the shared Gateway
domain: networking
host: cka-sim-control-plane
topics: [gateway-api, allowedroutes, namespaces, debugging]
weight: 4
---
## Context

Gateway `shared-gw` in Namespace `nw-gateway-allowed` is shared by app teams. HTTPRoute `team` in
Namespace `nw-gateway-allowed-team` points its `parentRefs` at `shared-gw`, for host
`team.nw-gateway-allowed.test`, but requests return 404 and the route isn't attached.

Namespaces allowed to use the shared Gateway carry the label `shared-gateway-access=true`.
`nw-gateway-allowed-team` already has it.

## Task

Change Gateway `shared-gw` so that its listener `http` accepts routes from every Namespace labelled
`shared-gateway-access=true`, and only those. Don't change the HTTPRoute.
