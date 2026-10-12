---
id: tr-qos
title: Make a Pod Guaranteed
domain: troubleshooting
host: cka-sim-control-plane
topics: [qos, resources, requests, limits]
weight: 4
---
## Context

Pod `payments` in Namespace `tr-qos` was evicted twice last week under memory pressure. It's
created from `/opt/course/tr-qos/payments.yaml`.

## Task

1. Write the QoS class of Pods `payments`, `search` and `nightly` (all in `tr-qos`) to
   `/opt/course/tr-qos/classes.txt`, one per line as `<pod>: <class>`.
2. Change `/opt/course/tr-qos/payments.yaml` so the Pod gets the QoS class `Guaranteed`, using
   `cpu: 200m` and `memory: 128Mi`, and recreate Pod `payments` from it.
