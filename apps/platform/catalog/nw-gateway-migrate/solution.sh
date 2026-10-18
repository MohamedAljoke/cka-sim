kubectl apply -f - <<'YAML'
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: shop-gw
  namespace: nw-gateway-migrate
spec:
  gatewayClassName: traefik
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
      - name: shop-tls
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: shop
  namespace: nw-gateway-migrate
spec:
  parentRefs:
  - name: shop-gw
  hostnames:
  - shop.nw-gateway-migrate.test
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /api
    backendRefs:
    - name: api
      port: 80
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: store
      port: 80
YAML
kubectl -n nw-gateway-migrate delete ingress shop
sleep 5
