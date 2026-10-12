---
id: ar-kubeconfig-cert
title: Read the certificate inside a kubeconfig
domain: architecture
host: cka-sim-control-plane
topics: [kubeconfig, certificates, base64, openssl]
weight: 3
---
## Context

`/opt/course/ar-kubeconfig-cert/auditor.kubeconfig` was handed out to an external auditor last
quarter. Nobody remembers which identity it carries.

## Task

Decode the client certificate embedded in that kubeconfig and write these lines to
`/opt/course/ar-kubeconfig-cert/identity.txt`:

```
user: <the certificate's CN>
groups: <all of its O values, comma-separated>
expires: <its expiry date, any format>
```
