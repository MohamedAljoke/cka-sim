fresh_ns nw-gateway-refgrant
fresh_ns nw-gateway-refgrant-data
web_app nw-gateway-refgrant-data catalog
gateway nw-gateway-refgrant gw
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: catalog
  namespace: nw-gateway-refgrant
spec:
  parentRefs:
  - name: gw
  hostnames:
  - catalog.nw-gateway-refgrant.test
  rules:
  - backendRefs:
    - name: catalog
      namespace: nw-gateway-refgrant-data
      port: 80
YAML
