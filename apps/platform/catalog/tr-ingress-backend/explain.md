Two backends were wrong. `describe` shows both, next to what each backend resolves to:

```bash
kubectl -n tr-ingress-backend describe ingress store
#   /      storefront:8080 (<error: ...>)
#   /api   store-api:80 (<error: services "store-api" not found>)
kubectl -n tr-ingress-backend get svc        # api 80/TCP, storefront 80/TCP
```

- **The port must be the Service's `port`,** not the container's or one you guess. `storefront`
  exposes 80, and the Ingress asked for 8080.
- **The name must be a Service in the same Namespace.** `store-api` doesn't exist, while `api` does.

Fix both in the Ingress (`kubectl edit ingress store`), not by adding a Service that matches the
mistake.

A quick way to tell what's broken:

| Symptom | Usual cause |
|---|---|
| 404 from the controller | no rule matches: host, path or pathType |
| 503 / 502 | the rule matches, but the Service, port or endpoints are wrong |
| your app's own 404 | routing works, the path reaches the app as something it doesn't serve (rewrite) |
