fresh_ns nw-gateway-rewrite
web_app nw-gateway-rewrite app
gateway nw-gateway-rewrite gw
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: legacy
  namespace: nw-gateway-rewrite
spec:
  parentRefs:
  - name: gw
  hostnames:
  - legacy.nw-gateway-rewrite.test
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /legacy
    backendRefs:
    - name: app
      port: 80
YAML
