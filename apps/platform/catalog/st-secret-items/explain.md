By default a Secret volume turns **every** key into a file named after the key, with mode `0644`.
`items` picks keys and renames them. `defaultMode` (whole volume) or per-item `mode` sets the
permissions:

```yaml
volumes:
- name: key
  secret:
    secretName: tls-bundle
    items:
    - key: tls.key
      path: server.key
      mode: 0400          # octal in YAML; in JSON you'd write the decimal 256
```

```sh
kubectl -n st-secret-items exec web -- ls -la /etc/ssl/private/
kubectl -n st-secret-items exec web -- stat -L -c '%a %U' /etc/ssl/private/server.key   # -L: the files are symlinks into ..data/
```

Leaving out the other keys is least privilege: if the app is compromised, the CA bundle and
certificate aren't sitting next to the key for no reason. A key listed in `items` that doesn't
exist in the Secret makes the mount fail, unless the volume sets `optional: true`.
