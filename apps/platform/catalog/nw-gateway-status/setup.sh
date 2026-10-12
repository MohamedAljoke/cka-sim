fresh_ns nw-gateway-status
fresh_course
web_app nw-gateway-status inventory
gateway nw-gateway-status gw
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: inventory
  namespace: nw-gateway-status
spec:
  parentRefs:
  - name: gw
  hostnames:
  - inventory.nw-gateway-status.test
  rules:
  - backendRefs:
    - name: inventry
      port: 80
YAML
