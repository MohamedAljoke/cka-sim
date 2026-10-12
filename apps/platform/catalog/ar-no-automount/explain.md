By default every Pod gets a projected volume `kube-api-access-*` with a token for its
ServiceAccount, mounted at `/var/run/secrets/kubernetes.io/serviceaccount/`. Anyone who gets code
running in the container can use it to call the API with that account's rights. Pods that never
call the API shouldn't carry it.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata: {name: batch-runner}
automountServiceAccountToken: false     # default for Pods using this SA
---
spec:
  serviceAccountName: batch-runner
  automountServiceAccountToken: false   # the Pod's setting wins over the SA's
```

```sh
kubectl -n ar-no-automount get pod runner -o yaml | grep -A3 volumes       # no kube-api-access volume
kubectl -n ar-no-automount exec runner -- ls /var/run/secrets/kubernetes.io/serviceaccount   # No such file
```

The Pod-level field overrides the ServiceAccount's in both directions, so a Pod can opt back in.
Every namespace's `default` ServiceAccount is a good candidate for `false`, since most apps run as
`default` without needing API access.
