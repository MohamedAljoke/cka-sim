kubectl apply -f - <<'YAML'
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
  - group: ""
    kind: Service
YAML
sleep 5
