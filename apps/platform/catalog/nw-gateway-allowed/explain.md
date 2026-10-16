The route's status says why it isn't attached:

```bash
kubectl -n nw-gateway-allowed-team describe httproute team   # Accepted False, NotAllowedByListeners
```

A listener's `allowedRoutes.namespaces.from` defaults to **`Same`**, so only routes in the Gateway's
own Namespace may attach. Open it to labelled Namespaces:

```yaml
  listeners:
  - name: http
    protocol: HTTP
    port: 80
    allowedRoutes:
      namespaces:
        from: Selector
        selector:
          matchLabels:
            shared-gateway-access: "true"
```

| `from` | Routes accepted from |
|---|---|
| `Same` (default) | the Gateway's Namespace |
| `All` | every Namespace |
| `Selector` | Namespaces matching `selector` |

Both sides must agree: the route names the Gateway in `parentRefs` (with `namespace:`), and the
Gateway's listener has to allow the route's Namespace. `All` would work too, but it lets any team
attach, so `Selector` is the least-privilege answer.
