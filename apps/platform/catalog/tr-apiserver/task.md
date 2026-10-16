---
id: tr-apiserver
title: API server down
domain: troubleshooting
host: cka-sim-control-plane
topics: [kube-apiserver, static-pods, etcd]
weight: 7
order: last
---
## Context

Since a change to the control plane this morning, every `kubectl` command on
`cka-sim-control-plane` fails with `connection refused`.

## Task

Find out why and fix it, so that the API server answers again and `kubectl get nodes` works.

Do not change the etcd configuration.
