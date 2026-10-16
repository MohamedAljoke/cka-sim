```bash
kubectl -n nw-ingress-rewrite annotate ingress app nginx.ingress.kubernetes.io/rewrite-target=/
curl -H 'Host: app.nw-ingress-rewrite.test' http://10.96.80.80/app     # web /
```

```yaml
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
```

- **Annotations are how an Ingress talks to its controller.** The Ingress API has no field for
  rewrites, so each controller reads its own annotations. The `nginx.ingress.kubernetes.io/` ones
  belong to ingress-nginx and to controllers that copy its behaviour, as this lab's does.
- **A fixed `rewrite-target: /` sends every match to `/`.** To keep the rest of the path
  (`/app/items` → `/items`), capture it with a regex:

```yaml
metadata:
  annotations:
    nginx.ingress.kubernetes.io/use-regex: "true"
    nginx.ingress.kubernetes.io/rewrite-target: /$2
spec:
  rules:
  - http:
      paths:
      - path: /app(/|$)(.*)
        pathType: ImplementationSpecific
```

Gateway API does the same with a portable `URLRewrite` filter instead of an annotation.
