fresh_ns nw-gateway-allowed
fresh_ns nw-gateway-allowed-team
fresh_ns nw-gateway-allowed-other
kubectl label namespace nw-gateway-allowed-team shared-gateway-access=true >/dev/null
gateway nw-gateway-allowed shared-gw
web_app nw-gateway-allowed-team app
web_app nw-gateway-allowed-other app
for ns in nw-gateway-allowed-team nw-gateway-allowed-other; do
  kubectl apply -f - >/dev/null <<YAML
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: team
  namespace: $ns
spec:
  parentRefs:
  - name: shared-gw
    namespace: nw-gateway-allowed
  hostnames:
  - ${ns#nw-gateway-allowed-}.nw-gateway-allowed.test
  rules:
  - backendRefs:
    - name: app
      port: 80
YAML
done
