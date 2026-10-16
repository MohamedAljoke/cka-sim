Each part of the Ingress has a place in Gateway API:

| Ingress | Gateway API |
|---|---|
| `ingressClassName: nginx` | Gateway `gatewayClassName: traefik` |
| `tls: [{hosts, secretName}]` | an HTTPS listener with `tls.certificateRefs` |
| `rules[].host` | HTTPRoute `hostnames` |
| `paths` with `pathType: Prefix` | `matches[].path` with `type: PathPrefix` |
| `backend.service.name/port` | `backendRefs[].name/port` |

```yaml
  listeners:
  - name: http
    protocol: HTTP
    port: 80
  - name: https
    protocol: HTTPS
    port: 443
    hostname: shop.nw-gateway-migrate.test
    tls:
      mode: Terminate
      certificateRefs:
      - name: shop-tls          # same Namespace as the Gateway
```

An HTTPRoute whose `parentRefs` has no `sectionName` attaches to every listener that allows it, so one
route serves both HTTP and HTTPS. Check before deleting the Ingress:

```bash
kubectl -n nw-gateway-migrate get gateway shop-gw          # PROGRAMMED True
kubectl -n nw-gateway-migrate describe httproute shop      # Accepted, ResolvedRefs
curl -k --resolve shop.nw-gateway-migrate.test:443:10.96.80.80 https://shop.nw-gateway-migrate.test/api/x
kubectl -n nw-gateway-migrate delete ingress shop
```

The `ingress2gateway` tool can write a first draft of this from existing Ingresses.
