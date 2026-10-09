---
id: tr-service
title: Service without backends
domain: troubleshooting
weight: 6
cluster: cka7491
host: cka7491-control-plane
---
## Context

The `web` Deployment in Namespace `shop` is running, but requests to its Service time out.

## Task

Fix Service `web` in Namespace `shop` so that `http://web.shop` (port `80`) serves the nginx
welcome page from inside the cluster.

Do not modify the Deployment `web`.
