The kubelet updates ConfigMap and Secret volumes in running Pods by itself, usually within a
minute. It writes the new files into a fresh hidden directory and swaps a `..data` symlink. A
**`subPath`** mount bind-mounts one file directly, so it keeps pointing at the old copy forever.
Pods with `subPath` mounts only see changes when they restart.

```yaml
volumeMounts:
- name: config
  mountPath: /etc/app            # the whole directory: /etc/app/settings.conf follows updates
# instead of
- name: config
  mountPath: /etc/app/settings.conf
  subPath: settings.conf         # frozen at Pod start
```

```sh
kubectl -n wl-subpath exec deploy/app -- ls -la /etc/app      # ..data -> ..2026_10_12_..., settings.conf -> ..data/settings.conf
```

The trade-off: mounting the directory hides anything else that was in `/etc/app` in the image.
That's why people reach for `subPath` in the first place. Mount at a dedicated directory and point
the app there if needed.

Env vars from ConfigMaps never update either. Only directory-mounted volumes do. Whether the app
notices a changed file is up to the app.
