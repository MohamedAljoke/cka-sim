```yaml
  rules:
  - backendRefs:
    - name: stable
      port: 80
      weight: 80
    - name: canary
      port: 80
      weight: 20
```

- **Weights are relative, not percentages.** `4` and `1` split traffic the same way as `80` and `20`.
  The default weight is `1`, and `0` takes a backend out of rotation without removing it.
- **Each request picks a backend at random by weight,** so 100 requests land near 80/20, not exactly
  on it.

Check the split:

```bash
for i in $(seq 50); do curl -s -H 'Host: shop.nw-gateway-canary.test' http://10.96.80.80/; done | sort | uniq -c
```

With an Ingress, the same needs controller-specific annotations (ingress-nginx's
`canary-weight`). In Gateway API it's part of the spec.
