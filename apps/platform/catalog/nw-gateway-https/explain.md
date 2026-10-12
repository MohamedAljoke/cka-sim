```yaml
# Gateway secure-gw: a second listener
  - name: https
    protocol: HTTPS
    port: 443
    hostname: secure.nw-gateway-https.test
    tls:
      mode: Terminate
      certificateRefs:
      - name: secure-tls
---
# HTTPRoute app: serve only on https
  parentRefs:
  - name: secure-gw
    sectionName: https
---
# HTTPRoute redirect: everything on http goes to https
  parentRefs:
  - name: secure-gw
    sectionName: http
  hostnames:
  - secure.nw-gateway-https.test
  rules:
  - filters:
    - type: RequestRedirect
      requestRedirect:
        scheme: https
        statusCode: 301
```

- **`sectionName` pins a route to one listener.** Without it, a route attaches to every listener that
  allows it. Then `app` would also answer plain HTTP, and the redirect would compete with it.
- **A rule with a `RequestRedirect` filter needs no `backendRefs`:** the Gateway answers the request
  itself. Leaving out `port` in the redirect lets the scheme's default (443) apply.
- **`mode: Terminate`** decrypts at the Gateway and sends plain HTTP to the Service. `Passthrough`
  would forward the TLS stream untouched (that needs a TLSRoute).

```bash
curl -i -H 'Host: secure.nw-gateway-https.test' http://10.96.80.80/x    # 301, Location: https://...
curl -k --resolve secure.nw-gateway-https.test:443:10.96.80.80 https://secure.nw-gateway-https.test/x
```
