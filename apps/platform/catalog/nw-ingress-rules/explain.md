`kubectl create ingress` writes the whole thing in one line. A `*` at the end of a path makes it
`Prefix`, and without it the path is `Exact`:

```bash
kubectl -n nw-ingress-rules create ingress shop --class=nginx \
  --rule='shop.nw-ingress-rules.test/api*=api:80' \
  --rule='shop.nw-ingress-rules.test/*=store:80'
```

```yaml
spec:
  ingressClassName: nginx
  rules:
  - host: shop.nw-ingress-rules.test
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service: {name: api, port: {number: 80}}
      - path: /
        pathType: Prefix
        backend:
          service: {name: store, port: {number: 80}}
```

- **The longest matching path wins,** so `/api/orders` goes to `api` and everything else to `store`.
- **The host comes from the `Host` header.** There's no DNS for `shop.nw-ingress-rules.test`, so send
  it yourself:

```bash
curl -H 'Host: shop.nw-ingress-rules.test' http://10.96.80.80/api/orders
# or, closer to a real client:
curl --resolve shop.nw-ingress-rules.test:80:10.96.80.80 http://shop.nw-ingress-rules.test/api/orders
```

`kubectl get ingress` shows the controller's address under ADDRESS once it has picked the Ingress up.
