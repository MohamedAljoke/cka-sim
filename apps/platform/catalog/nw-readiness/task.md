---
id: nw-readiness
title: Service with no endpoints
domain: networking
host: cka-sim-control-plane
topics: [services, endpoints, readiness-probes]
weight: 5
---
## Context

Deployment `web` in Namespace `nw-readiness` is running, but Service `web` has no endpoints and
requests to it fail.

## Task

Fix it, so that `http://web.nw-readiness` serves the nginx welcome page from inside the cluster.

- Do not modify Service `web`.
- The container must keep an HTTP readiness probe.
