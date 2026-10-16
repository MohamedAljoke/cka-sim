```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: web-gw
  namespace: nw-gateway-routes
spec:
  gatewayClassName: traefik
  listeners:
  - name: http
    protocol: HTTP
    port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: web-route
  namespace: nw-gateway-routes
spec:
  parentRefs:
  - name: web-gw
  hostnames:
  - app.nw-gateway-routes.test
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /api
    backendRefs:
    - name: api
      port: 80
  - backendRefs:          # no matches = path prefix /
    - name: web
      port: 80
```

The three objects split the work by role:

- **GatewayClass** (`traefik`) is cluster-scoped: which controller implements Gateways. Run `kubectl get
  gatewayclass` to see what's available.
- **Gateway** is the entry point: listeners with protocol, port and optionally hostname and TLS.
  Platform teams own it.
- **HTTPRoute** holds the routing: it attaches to a Gateway through `parentRefs`. App teams own it.

Read the status before you test. Both should say True:

```bash
kubectl -n nw-gateway-routes get gateway web-gw              # PROGRAMMED True, ADDRESS 10.96.80.80
kubectl -n nw-gateway-routes describe httproute web-route    # Accepted, ResolvedRefs
curl -H 'Host: app.nw-gateway-routes.test' http://10.96.80.80/api/users
```

The most specific match wins, so `/api/...` goes to `api` whatever the order of the rules.
