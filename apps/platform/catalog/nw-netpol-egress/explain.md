```yaml
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
```

- **Write `policyTypes: [Egress]`.** Kubernetes would infer it from the `egress` rules, but listing
  it means `backend`'s ingress is left alone. If you also listed `Ingress`, nothing could reach
  `backend`.
- **`to` and `ports` in one rule are ANDed:** a Pod labelled `app=db` *and* port 5432. Splitting
  them into two rules would allow any port on `db` and port 5432 everywhere.
- **The policy selects only `app=backend`,** so `frontend` keeps full access.

In practice you'd usually add a DNS rule too, as in the default-deny task. Without it, `backend`
can't resolve the `db` Service's name.
