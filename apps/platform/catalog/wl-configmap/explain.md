A ConfigMap is a set of keys. `--from-literal=K=V` adds a key with a value, and `--from-file=path`
adds a key named after the file, holding the file's content. A Pod consumes keys in two ways:

```yaml
env:
- name: LOG_LEVEL
  valueFrom:
    configMapKeyRef: {name: app-config, key: LOG_LEVEL}   # one key -> one env var
volumes:
- name: config
  configMap:
    name: app-config
    items:                                               # only these keys become files
    - {key: app.properties, path: app.properties}
```

Without `items`, every key becomes a file in the mount directory, including `LOG_LEVEL`.
`envFrom` with `configMapRef` turns every key into an env var at once.

```sh
kubectl -n wl-configmap create configmap app-config --from-literal=LOG_LEVEL=debug --from-file=/opt/course/wl-configmap/app.properties
kubectl -n wl-configmap run app --image=busybox:1.36 --dry-run=client -o yaml --command -- sleep 1d > app.yaml   # then add env + volume
kubectl -n wl-configmap exec app -- cat /etc/app/app.properties
kubectl -n wl-configmap exec app -- printenv LOG_LEVEL
```

Mounted files update when the ConfigMap changes, after up to about a minute. Env vars are read only
when the container starts, so they never change.
