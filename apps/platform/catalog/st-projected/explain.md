A **projected** volume merges several sources into one directory: `configMap`, `secret`,
`downwardAPI`, `serviceAccountToken` and `clusterTrustBundle`. Each source lists its files with
`items` or `path`.

```yaml
volumes:
- name: all-in-one
  projected:
    sources:
    - configMap: {name: app-config, items: [{key: settings, path: settings}]}
    - secret:    {name: app-secret, items: [{key: token, path: token}]}
    - downwardAPI:
        items: [{path: labels, fieldRef: {fieldPath: metadata.labels}}]
    - serviceAccountToken: {audience: vault, expirationSeconds: 3600, path: vault-token}
```

The `serviceAccountToken` source is the interesting one. The kubelet requests a token through the
TokenRequest API for **this Pod**, with your audience and lifetime, and **refreshes** it before it
expires. A service like Vault can verify it and accept only tokens meant for `vault`. It's how
every Pod already gets its API token: look for the `kube-api-access-*` projected volume in any
Pod's spec.

```sh
kubectl -n st-projected exec app -- ls /etc/app
kubectl -n st-projected exec app -- cat /etc/app/vault-token | cut -d. -f2 | base64 -d 2>/dev/null
```
