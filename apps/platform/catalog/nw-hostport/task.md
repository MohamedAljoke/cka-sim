---
id: nw-hostport
title: Expose a container with hostPort
domain: networking
host: cka-sim-control-plane
topics: [hostport, pods, nodeport]
weight: 3
---
## Task

In Namespace `nw-hostport`, create Pod `web` (image `nginx:1.27-alpine`) on node
`cka-sim-worker2` that publishes its container port `80` on port `8081` of that node, the way
`docker run -p 8081:80` would. Keep the Pod on its own network (no `hostNetwork`) and create no
Service.

`http://<IP of cka-sim-worker2>:8081` must serve the nginx welcome page.
