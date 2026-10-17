fresh_ns nw-netpol-and-or
fresh_ns nw-netpol-mon
kubectl label namespace nw-netpol-mon team=monitoring >/dev/null
kubectl -n nw-netpol-and-or run api --image=busybox:1.36 --labels=app=api --command -- sh -c 'httpd -p 8080; sleep 1d' >/dev/null
kubectl -n nw-netpol-and-or run prometheus --image=busybox:1.36 --labels=app=prometheus --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-mon run prometheus --image=busybox:1.36 --labels=app=prometheus --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-mon run grafana --image=busybox:1.36 --labels=app=grafana --command -- sleep 1d >/dev/null
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-from-monitoring
  namespace: nw-netpol-and-or
spec:
  podSelector:
    matchLabels:
      app: api
  policyTypes: [Ingress]
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          team: monitoring
    - podSelector:
        matchLabels:
          app: prometheus
    ports:
    - protocol: TCP
      port: 8080
YAML
kubectl -n nw-netpol-and-or wait --for=condition=Ready pod --all --timeout=120s >/dev/null
kubectl -n nw-netpol-mon wait --for=condition=Ready pod --all --timeout=120s >/dev/null
