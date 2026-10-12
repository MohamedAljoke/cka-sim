A kubeconfig fails in a handful of ways, and `config view` shows most of them (secrets are
redacted unless you add `--raw`):

```sh
KC=/opt/course/ar-broken-kubeconfig/dev.kubeconfig
kubectl --kubeconfig $KC config view
kubectl --kubeconfig $KC config get-contexts
kubectl --kubeconfig $KC get nodes -v=6          # shows which URL and credentials kubectl actually uses
```

| Symptom | Cause |
|---|---|
| `The connection to the server ... was refused` | wrong host or port (the API server listens on 6443) |
| `current-context is not set` / falls back to `localhost:8080` | no `current-context` |
| `user "dev-admin" not found`-style errors, or an anonymous 403 | the context references a user (or cluster) entry that doesn't exist |
| `x509: certificate signed by unknown authority` | wrong or missing `certificate-authority-data` |
| `Unauthorized` | bad or expired credentials |

```sh
kubectl --kubeconfig $KC config set-cluster cka-sim --server=https://cka-sim-control-plane:6443
kubectl --kubeconfig $KC config set-context dev --user=admin
kubectl --kubeconfig $KC config use-context dev
```

`set-cluster` and `set-context` only change the fields you pass and keep the rest, including the
embedded certificates.
