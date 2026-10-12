---
id: ar-ca-map
title: Which CA signs which certificate
domain: architecture
host: cka-sim-control-plane
topics: [certificates, pki, kubeadm, etcd]
weight: 4
---
## Context

A kubeadm cluster has more than one certificate authority under `/etc/kubernetes/pki/`.

## Task

For each certificate below, write the path of the CA certificate that signed it to
`/opt/course/ar-ca-map/ca.txt`, one line each:

```
apiserver.crt: <CA file>
apiserver-etcd-client.crt: <CA file>
front-proxy-client.crt: <CA file>
etcd/server.crt: <CA file>
```

All paths are under `/etc/kubernetes/pki/`; write them in full.
