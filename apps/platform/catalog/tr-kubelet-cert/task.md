---
id: tr-kubelet-cert
title: Kubelet client certificate rejected
domain: troubleshooting
host: cka-sim-worker
topics: [kubelet, certificates, tls, nodes]
weight: 7
order: last
---
## Context

Node `cka-sim-worker` went `NotReady` right after someone rotated its kubelet client certificate
by hand. The kubelet service itself is running.

## Task

Find out why the API server rejects the kubelet, and fix it so that node `cka-sim-worker` is
`Ready` again. The kubelet must use a client certificate signed by the cluster CA
(`/etc/kubernetes/pki/ca.crt`).
