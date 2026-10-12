---
id: nw-hostnetwork
title: Pod on the host's network
domain: networking
host: cka-sim-control-plane
topics: [hostnetwork, pods, network-namespaces]
weight: 3
---
## Task

In Namespace `nw-hostnetwork`, create Pod `node-agent` (image `busybox:1.36`, command
`sh -c 'mkdir -p /www && echo agent > /www/index.html && httpd -f -p 18080 -h /www'`) on node
`cka-sim-worker2` that uses the **node's** network namespace instead of its own.

`http://<IP of cka-sim-worker2>:18080` must then answer `agent`, with no Service involved.
