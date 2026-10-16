---
id: nw-ingress-tls
title: Serve an Ingress over TLS
domain: networking
host: cka-sim-control-plane
topics: [ingress, tls, secrets]
weight: 5
---
## Context

Ingress `secure` in Namespace `nw-ingress-tls` serves host `secure.nw-ingress-tls.test` over plain
HTTP (IngressClass `nginx`, controller at `10.96.80.80`). A certificate and key for that host are in
`/opt/course/nw-ingress-tls/tls.crt` and `/opt/course/nw-ingress-tls/tls.key`.

## Task

1. Create a TLS Secret `secure-tls` in Namespace `nw-ingress-tls` from those two files.
2. Change Ingress `secure` so that `https://secure.nw-ingress-tls.test` is served with that
   certificate.
