fresh_ns nw-gateway-canary
web_app nw-gateway-canary stable
web_app nw-gateway-canary canary
gateway nw-gateway-canary gw
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: shop
  namespace: nw-gateway-canary
spec:
  parentRefs:
  - name: gw
  hostnames:
  - shop.nw-gateway-canary.test
  rules:
  - backendRefs:
    - name: stable
      port: 80
YAML
