---
id: tr-besteffort
title: Find Pods without resource requests
domain: troubleshooting
host: cka-sim-control-plane
topics: [qos, resources, jsonpath, custom-columns]
weight: 3
---
## Context

Pods without any resource requests or limits are the first to be evicted when a node runs short of
memory. The team wants a list of them in their two namespaces, `tr-besteffort-a` and
`tr-besteffort-b`.

## Task

Write every Pod in those two namespaces whose QoS class is `BestEffort` to
`/opt/course/tr-besteffort/pods.txt`, one per line as `<namespace>/<name>`.
