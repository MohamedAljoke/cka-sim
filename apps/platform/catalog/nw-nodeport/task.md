---
id: nw-nodeport
title: Expose a Deployment on every node
domain: networking
host: cka-sim-control-plane
topics: [services, nodeport]
weight: 3
---
## Context

Deployment `web` in Namespace `nw-nodeport` runs nginx on port `80`.

## Task

Create a NodePort Service `web` that makes the Deployment reachable on port `30080` of every
node, so that `curl http://<any node IP>:30080` returns the nginx welcome page.
