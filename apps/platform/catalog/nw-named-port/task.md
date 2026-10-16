---
id: nw-named-port
title: Target a port by name
domain: networking
host: cka-sim-control-plane
topics: [services, ports, deployments]
weight: 4
---
## Context

Deployment `api` in Namespace `nw-named-port` serves HTTP on the port in its env var `PORT`
(`8080`), and its container port `8080` is named `http`. The team is about to move it to `9090`.

## Task

1. Create Service `api` on port `80` that targets the container port **by its name** `http`, not
   by number.
2. Move the app to port `9090`: set `PORT=9090` and change the container port named `http` to
   `9090`. Do not touch Service `api` again.

`http://api.nw-named-port` must serve the app from Pod `client` afterwards.
