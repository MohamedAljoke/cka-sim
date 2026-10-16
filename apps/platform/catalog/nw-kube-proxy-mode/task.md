---
id: nw-kube-proxy-mode
title: How kube-proxy routes a Service
domain: networking
host: cka-sim-control-plane
topics: [kube-proxy, iptables, services]
weight: 4
---
## Context

Namespace `nw-kube-proxy-mode` has Service `api` with 2 endpoints.

## Task

Write these lines to `/opt/course/nw-kube-proxy-mode/proxy.txt`:

```
mode: <the proxy mode kube-proxy runs in>
chain: <the name of the KUBE-SVC-... chain kube-proxy created for Service api, on this node>
endpoints: <how many KUBE-SEP-... chains that Service chain jumps to>
```
