kubectl apply -f - <<'YAML'
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-egress
  namespace: nw-netpol-egress
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes: [Egress]
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: db
    ports:
    - protocol: TCP
      port: 5432
YAML
sleep 5
