fresh_ns nw-netpol-pick
fresh_ns nw-netpol-pick-other
fresh_course
kubectl -n nw-netpol-pick run db --image=busybox:1.36 --labels=app=db --command -- sh -c 'httpd -p 5432; httpd -p 9000; sleep 1d' >/dev/null
kubectl -n nw-netpol-pick run backend --image=busybox:1.36 --labels=app=backend --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-pick run frontend --image=busybox:1.36 --labels=app=frontend --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-pick-other run backend --image=busybox:1.36 --labels=app=backend --command -- sleep 1d >/dev/null
policy() {
  cat > "$COURSE/policy-$1.yaml" <<YAML
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: db-$1
  namespace: nw-netpol-pick
spec:
  podSelector:
    matchLabels:
      app: db
  policyTypes: [Ingress]
  ingress:
$(cat)
YAML
}
policy a <<'YAML'
  - from:
    - podSelector:
        matchLabels:
          app: backend
YAML
policy b <<'YAML'
  - from:
    - namespaceSelector: {}
      podSelector:
        matchLabels:
          app: backend
    ports:
    - protocol: TCP
      port: 5432
YAML
policy c <<'YAML'
  - from:
    - podSelector:
        matchLabels:
          app: backend
    ports:
    - protocol: TCP
      port: 5432
YAML
policy d <<'YAML'
  - from:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 5432
YAML
kubectl -n nw-netpol-pick wait --for=condition=Ready pod --all --timeout=120s >/dev/null
kubectl -n nw-netpol-pick-other wait --for=condition=Ready pod --all --timeout=120s >/dev/null
