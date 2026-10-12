---
id: ar-auth-config
title: How the API server checks requests
domain: architecture
host: cka-sim-control-plane
topics: [authentication, authorization, kube-apiserver]
weight: 3
---
## Task

Find out how this cluster's API server authenticates and authorizes requests, and write these
lines to `/opt/course/ar-auth-config/auth.txt`:

```
client-ca-file: <file of the CA that client certificates must be signed by>
anonymous-auth: <true or false: are unauthenticated requests accepted as system:anonymous>
authorization-mode: <the authorizers, in order, comma-separated as configured>
```
