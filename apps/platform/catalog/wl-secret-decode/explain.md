`kubectl get secret -o yaml` shows each value base64-encoded. Pull out one key with jsonpath and
decode it:

```sh
kubectl -n wl-secret-decode get secret api-token -o jsonpath='{.data.token}' | base64 -d > /opt/course/wl-secret-decode/token.txt
kubectl -n wl-secret-decode get secret api-token -o go-template='{{.data.token | base64decode}}'   # same, no pipe
```

The traps: copying the encoded value into the file, or `kubectl describe secret`, which shows only
the size of each value. A key with a dot in its name needs escaping in jsonpath:
`{.data.tls\.crt}`.

Anyone who can `get` a Secret can decode it this way. Base64 only makes binary data safe to put in
YAML. Protect Secrets with RBAC and, on the API server, encryption at rest.
