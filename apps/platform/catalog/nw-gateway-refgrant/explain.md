```bash
kubectl -n nw-gateway-refgrant describe httproute catalog   # ResolvedRefs False, RefNotPermitted
```

A reference across Namespaces needs consent from the **target** Namespace, so the ReferenceGrant lives
next to the Service:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: ReferenceGrant
metadata:
  name: allow-web-routes
  namespace: nw-gateway-refgrant-data
spec:
  from:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    namespace: nw-gateway-refgrant
  to:
  - group: ""          # core API group, where Services live
    kind: Service
```

- **`from` names who may refer** (kind and Namespace), **`to` names what they may refer to.** Add
  `name:` under `to` to allow one Service only.
- **The same object governs Gateways reading TLS Secrets** from another Namespace (`from` kind
  `Gateway`, `to` kind `Secret`).
- `kubectl api-resources | grep referencegrant` shows which API version your cluster serves.

Ingress has no such mechanism: its backends must be in its own Namespace.
