```yaml
spec:
  ingressClassName: nginx
  defaultBackend:
    service:
      name: fallback
      port:
        number: 80
  rules:
  - host: www.nw-ingress-default.test
    ...
```

- **`spec.defaultBackend` takes every request that matches no rule:** unknown hosts, and paths no
  rule covers. Without it the controller answers with its own 404.
- **It's a field of the Ingress, not of a rule,** so it sits next to `rules`, not inside one.
- `kubectl create ingress` can set it too: `--default-backend=fallback:80`.

```bash
curl -H 'Host: anything.test' http://10.96.80.80/x               # fallback /x
curl -H 'Host: www.nw-ingress-default.test' http://10.96.80.80/  # site /
kubectl -n nw-ingress-default describe ingress site              # Default backend: fallback:80
```

A default backend acts for the whole controller, so in a shared cluster only one Ingress should set it.
