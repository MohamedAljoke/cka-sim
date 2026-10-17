fresh_ns nw-gateway-header
web_app nw-gateway-header v1
web_app nw-gateway-header v2
gateway nw-gateway-header gw
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: app
  namespace: nw-gateway-header
spec:
  parentRefs:
  - name: gw
  hostnames:
  - app.nw-gateway-header.test
  rules:
  - backendRefs:
    - name: v1
      port: 80
YAML
