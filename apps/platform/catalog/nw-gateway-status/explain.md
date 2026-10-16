A route's status is per parent Gateway, under `status.parents[]`:

```bash
kubectl -n nw-gateway-status describe httproute inventory
kubectl -n nw-gateway-status get httproute inventory \
  -o jsonpath='{range .status.parents[0].conditions[*]}{.type}={.status} {.reason}{"\n"}{end}'
# Accepted=True Accepted
# ResolvedRefs=False BackendNotFound
```

The backend was `inventry`, a typo:

```bash
kubectl -n nw-gateway-status patch httproute inventory --type=json \
  -p '[{"op":"replace","path":"/spec/rules/0/backendRefs/0/name","value":"inventory"}]'
```

The conditions to know:

| Object | Condition | False means |
|---|---|---|
| Gateway | `Accepted` | the GatewayClass or a listener is invalid |
| Gateway | `Programmed` | the controller hasn't configured the data plane (yet) |
| HTTPRoute | `Accepted` | it can't attach: `NotAllowedByListeners`, `NoMatchingParent`, `NoMatchingListenerHostname` |
| HTTPRoute | `ResolvedRefs` | a backend is wrong: `BackendNotFound`, `RefNotPermitted`, `InvalidKind` |

An attached route with a missing backend answers **500**, not 404. The route exists, so the request
matches, but there's nowhere to send it.
