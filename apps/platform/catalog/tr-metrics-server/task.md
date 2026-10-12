---
id: tr-metrics-server
title: Install metrics-server
domain: troubleshooting
host: cka-sim-control-plane
topics: [metrics-server, kubectl-top, tls]
weight: 5
order: last
---
## Context

`kubectl top nodes` fails with `Metrics API not available`: nobody installed metrics-server. The
upstream manifest was downloaded to `/opt/course/tr-metrics-server/components.yaml`.

## Task

Install metrics-server from `/opt/course/tr-metrics-server/components.yaml` into Namespace
`kube-system`, so that `kubectl top nodes` shows the usage of every node.

The kubelets of this cluster serve self-signed certificates, without IP SANs.
