```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-allow
  namespace: nw-netpol-ingress
spec:
  podSelector:
    matchLabels:
      app: api
  policyTypes: [Ingress]
  ingress:
  - from:
    - podSelector:
        matchLabels:
          role: frontend
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: nw-netpol-monitor
    ports:
    - protocol: TCP
      port: 8080
```

- **Once a policy selects `api` for ingress, only what some policy allows gets in.** So `worker`,
  `stranger` and port `9090` are blocked without a single deny rule.
- **The two `from` entries are ORed.** A `podSelector` on its own matches Pods in the policy's own
  Namespace only. That's why `stranger`, labelled `role=frontend` in another Namespace, stays out.
- **`kubernetes.io/metadata.name`** is a label Kubernetes puts on every Namespace, so you can select
  one by name without labelling it yourself.
- **`ports` applies to the whole rule**, to both sources.

Check it from the client Pods: `kubectl -n nw-netpol-ingress exec frontend -- nc -z -w 2 <api IP> 8080`.
