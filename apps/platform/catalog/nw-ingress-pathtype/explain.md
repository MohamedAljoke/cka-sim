The path was `pathType: Exact`, so only `/docs` itself matched:

```bash
kubectl -n nw-ingress-pathtype get ingress docs -o jsonpath='{.spec.rules[0].http.paths[0]}'
kubectl -n nw-ingress-pathtype edit ingress docs    # pathType: Exact -> Prefix
```

| pathType | `/docs` matches |
|---|---|
| `Exact` | `/docs` only |
| `Prefix` | `/docs`, `/docs/`, `/docs/guide`, but **not** `/documents` |
| `ImplementationSpecific` | whatever the controller decides |

`Prefix` compares whole path elements split on `/`, not characters. That's why `/documents` stays out.

`kubectl create ingress` picks the type from the path: `/docs` makes it `Exact`, and `/docs*` makes it
`Prefix`.
