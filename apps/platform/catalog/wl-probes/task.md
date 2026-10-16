---
id: wl-probes
title: Startup, liveness and readiness probes
domain: workloads
host: cka-sim-control-plane
topics: [probes, self-healing, deployments]
weight: 5
---
## Task

Deployment `api` in Namespace `wl-probes` runs nginx on port `80`. Add three probes to its
container:

1. A **startup** probe: HTTP `GET /` on port `80`, every `2` seconds, up to `30` failures.
2. A **liveness** probe: a TCP check on port `80`, every `10` seconds.
3. A **readiness** probe: the command `cat /usr/share/nginx/html/index.html`, every `5` seconds.

All Pods of `api` must be Ready afterwards.
