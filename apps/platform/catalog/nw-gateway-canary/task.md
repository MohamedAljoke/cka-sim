---
id: nw-gateway-canary
title: Send 20% of traffic to the canary
domain: networking
host: cka-sim-control-plane
topics: [gateway-api, httproute, traffic-splitting]
weight: 4
---
## Context

HTTPRoute `shop` in Namespace `nw-gateway-canary` sends every request for
`shop.nw-gateway-canary.test` to Service `stable` (port `80`) through Gateway `gw`. A new version
runs as Service `canary` (port `80`).

## Task

Change HTTPRoute `shop` so that about **80%** of requests go to `stable` and **20%** to `canary`.
Use backend weights in a single rule.
