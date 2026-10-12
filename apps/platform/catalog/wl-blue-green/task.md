---
id: wl-blue-green
title: Blue/green switch
domain: workloads
host: cka-sim-control-plane
topics: [deployments, services, labels, releases]
weight: 4
---
## Context

Namespace `wl-blue-green` runs version blue (Deployment `web-blue`) and version green (Deployment
`web-green`) side by side. Service `web` sends all traffic to blue.

## Task

1. Switch Service `web` so that all its traffic goes to green, with no moment without backends.
2. After the switch, scale `web-blue` down to `0` replicas. Keep the Deployment.
