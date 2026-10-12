Every API request passes three stages: **authentication** (who are you?), **authorization** (may
you do this?), then **admission** (should this object be changed or rejected?). Their settings are
API server flags, so on kubeadm clusters they live in the static Pod manifest:

```sh
grep -E 'client-ca-file|anonymous-auth|authorization-mode|oidc|token-auth' /etc/kubernetes/manifests/kube-apiserver.yaml
```

- **Authentication**: `--client-ca-file` enables X.509 client certificates (CN = user,
  O = groups). ServiceAccount tokens, bootstrap tokens and OIDC are other methods. With
  `--anonymous-auth` at its default `true`, requests with no credentials become `system:anonymous`.
  They are still subject to authorization, and RBAC grants them almost nothing (`/healthz`,
  `/version`...).
- **Authorization**: `--authorization-mode=Node,RBAC` asks each authorizer in turn until one allows
  the request. **Node** lets kubelets touch only their own node's objects, and **RBAC** handles
  everyone else. `AlwaysAllow` would turn security off, so never use it outside a lab.

```sh
kubectl get --raw /version --as=system:anonymous      # allowed by the default system:public-info-viewer
curl -k https://localhost:6443/api/v1/pods            # 403 for system:anonymous: authenticated, not authorized
```
