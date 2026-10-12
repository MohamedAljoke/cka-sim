---
id: nw-gateway-status
title: Read why the route doesn't work
domain: networking
host: cka-sim-control-plane
topics: [gateway-api, status, debugging]
weight: 4
---
## Context

HTTPRoute `inventory` in Namespace `nw-gateway-status` (host `inventory.nw-gateway-status.test`,
Gateway `gw`) should send traffic to Service `inventory` on port `80`, but requests fail.

## Task

1. Find the condition on HTTPRoute `inventory` whose status is `False`. Write its type and reason,
   separated by a space, to `/opt/course/nw-gateway-status/condition.txt`, for example
   `Accepted NotAllowedByListeners`.
2. Fix the HTTPRoute so that requests reach Service `inventory`.
