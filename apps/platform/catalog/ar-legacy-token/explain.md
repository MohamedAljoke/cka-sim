Since v1.24, ServiceAccounts no longer get a token Secret automatically. Tokens are short-lived
and requested on demand (`kubectl create token`, or the projected volume in Pods). For a system
that can't renew tokens, you can still ask for an old-style, **non-expiring** token by creating
the Secret yourself:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deployer-token
  annotations:
    kubernetes.io/service-account.name: deployer     # which ServiceAccount
type: kubernetes.io/service-account-token
```

The token controller fills in `data.token`, `ca.crt` and `namespace` within a second.

```sh
kubectl -n ar-legacy-token get secret deployer-token -o jsonpath='{.data.token}' | base64 -d > token
kubectl --token="$(cat token)" auth whoami
```

| | `kubectl create token` | token Secret |
|---|---|---|
| Lifetime | 1h by default, chosen with `--duration` | never expires |
| Stored in the cluster | no | yes, in a Secret anyone with `get secrets` can read |
| Revoke | wait for expiry, or delete the ServiceAccount | delete the Secret |

Use the Secret only when you must, and rotate it by deleting and recreating it. The control plane
also flags unused legacy tokens (`kubernetes.io/legacy-token-invalid-since`).
