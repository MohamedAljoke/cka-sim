```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny
  namespace: nw-netpol-lockdown
spec:
  podSelector: {}
  policyTypes: [Ingress, Egress]
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-dns
  namespace: nw-netpol-lockdown
spec:
  podSelector: {}
  policyTypes: [Egress]
  egress:
  - ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
```

- **`podSelector: {}` selects every Pod in the Namespace.** With no `ingress` or `egress` rules,
  it allows nothing in either direction.
- **`policyTypes` must list `Egress`.** A policy with only `ingress` rules defaults to
  `[Ingress]`, so without the explicit list egress would stay open.
- **The DNS rule has no `to`, only `ports`.** That allows port 53 to any destination, so CoreDNS
  is reached even though its Pods live in `kube-system`. To lock it down further, add
  `to: [{namespaceSelector: {matchLabels: {kubernetes.io/metadata.name: kube-system}}}]`.
- **Allow both UDP and TCP.** Large answers and some resolvers fall back to TCP.

Policies add up, so `allow-dns` opens a hole in `default-deny` without editing it. Test it with
`kubectl -n nw-netpol-lockdown exec app -- nslookup kubernetes.default`, which works, and
`nc -z -w 2 kubernetes.default 443`, which times out.
