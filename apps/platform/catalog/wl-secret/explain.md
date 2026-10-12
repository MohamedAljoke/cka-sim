Secrets work like ConfigMaps, with a few differences: values are stored base64-encoded (which is
encoding, not encryption), they're kept in memory on the node (`tmpfs`), and RBAC usually guards
them more tightly.

```sh
kubectl -n wl-secret create secret generic db-creds --from-literal=username=admin --from-literal='password=S3cr3t!'
```

Quote values with `!`, `$` or spaces, or the shell changes them before kubectl sees them. In the
Pod, `envFrom` turns every key into an env var. In a volume, each key becomes a file:

```yaml
envFrom:
- secretRef: {name: db-creds}
volumeMounts:
- {name: creds, mountPath: /etc/db-creds, readOnly: true}
volumes:
- name: creds
  secret: {secretName: db-creds}      # note: secretName, where a ConfigMap uses name
```

`readOnly: true` belongs on the **volumeMount**. The volume itself has no such field for Secrets.
Env vars leak more easily than files (child processes, crash dumps, `kubectl describe` of some
tools), which is why files are often preferred for passwords.
