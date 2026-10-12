Helm marks every object it creates with two annotations and a label:

```sh
kubectl -n ar-helm-owner get deployment checkout -o yaml | grep -A3 annotations
```

```yaml
  annotations:
    meta.helm.sh/release-name: payments
    meta.helm.sh/release-namespace: ar-helm-owner
  labels:
    app.kubernetes.io/managed-by: Helm
```

```sh
kubectl -n ar-helm-owner get deployment checkout \
  -o jsonpath='{.metadata.annotations.meta\.helm\.sh/release-name}'
```

The dots inside the key must be escaped in JSONPath. Here every release set `fullnameOverride`, so
the names give nothing away. Edits made by hand are undone by the next `helm upgrade` of
`payments`.

`helm ls` shows only the current Namespace. `-A` (`--all-namespaces`) shows all, and `-a` adds the
releases that aren't `deployed`:

```sh
helm ls -A
helm ls -A -a
```
