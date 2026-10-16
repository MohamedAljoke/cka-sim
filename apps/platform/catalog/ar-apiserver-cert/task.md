---
id: ar-apiserver-cert
title: Read the API server certificate
domain: architecture
host: cka-sim-control-plane
topics: [certificates, tls, openssl, kube-apiserver]
weight: 3
---
## Context

A new load balancer will reach the API server as `k8s.cka.internal`. Before switching, the team
needs to know whether the current serving certificate covers that name.

## Task

Inspect the kube-apiserver's serving certificate and write these lines to
`/opt/course/ar-apiserver-cert/cert.txt`:

```
issuer: <the issuer's CN>
expires: <the expiry date, any format>
dns-names: <all DNS names in the Subject Alternative Names, comma-separated>
k8s.cka.internal: <yes or no: is that name covered>
```
