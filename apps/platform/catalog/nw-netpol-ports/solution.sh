kubectl apply -f - <<'YAML'
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: server-ports
  namespace: nw-netpol-ports
spec:
  podSelector:
    matchLabels:
      app: server
  policyTypes: [Ingress]
  ingress:
  - ports:
    - protocol: TCP
      port: 8000
      endPort: 8010
    - protocol: TCP
      port: metrics
YAML
sleep 5
