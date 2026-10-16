---
id: nw-gateway-https
title: Add HTTPS and redirect HTTP to it
domain: networking
host: cka-sim-control-plane
topics: [gateway-api, tls, redirect]
weight: 6
---
## Context

Gateway `secure-gw` in Namespace `nw-gateway-https` has one listener, `http` (port `80`). HTTPRoute
`app` serves `secure.nw-gateway-https.test` from Service `app` (port `80`). TLS Secret `secure-tls`
in the same Namespace holds a certificate for that host.

## Task

1. Add a listener `https` to `secure-gw`: protocol `HTTPS`, port `443`, hostname
   `secure.nw-gateway-https.test`, terminating TLS with Secret `secure-tls`.
2. Make HTTPRoute `app` attach only to the `https` listener.
3. Create an HTTPRoute `redirect`, attached to the `http` listener for the same host, that answers
   every request with a `301` redirect to `https`.
