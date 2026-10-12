`helm upgrade` decides where to start from:

- **No `--set`/`-f` at all:** it reuses the last release's values.
- **Any `--set` or `-f`:** it starts again from the chart's defaults plus only what you passed.
  This is the behaviour of `--reset-values`. Your colleague's upgrade kept `pullPolicy` and
  silently dropped `replicaCount=3` and `service.type=NodePort`.

`--reuse-values` merges the new flags into the last release's values instead:

```sh
helm -n ar-helm-reuse get values portal
helm -n ar-helm-reuse upgrade portal web --repo http://cka-sim-charts --version 1.1.0 \
  --reuse-values --set replicaCount=3 --set service.type=NodePort
```

Rolling back to revision 1 would bring back 3 replicas and the NodePort, but lose `pullPolicy`.
Repeating every value with `--set` works too, but you have to know them all. That's why teams keep
a values file and always pass `-f values.yaml`.

`--reset-then-reuse-values` (Helm 3.14+) is the middle way: chart defaults from the new chart
version, plus the old user-supplied values. It's useful when an upgrade adds new defaults that
`--reuse-values` would skip.
