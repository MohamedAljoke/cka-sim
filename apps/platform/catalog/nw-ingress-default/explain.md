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

- **`spec.defaultBackend` takes the requests that match none of the rules' paths,** here everything on
  the host outside `/shop`. Without it the controller answers with its own 404.
- **It's a field of the Ingress, not of a rule,** so it sits next to `rules`, not inside one.
- `kubectl create ingress` can set it too: `--default-backend=fallback:80`.

```bash
curl -H 'Host: www.nw-ingress-default.test' http://10.96.80.80/blog       # fallback /blog
curl -H 'Host: www.nw-ingress-default.test' http://10.96.80.80/shop/cart  # site /shop/cart
kubectl -n nw-ingress-default describe ingress site              # Default backend: fallback:80
```

Controllers differ on other hosts. An Ingress with a `defaultBackend` and no rules at all catches every
host the controller doesn't know; one with rules, like this, only catches its own hosts' leftovers.
