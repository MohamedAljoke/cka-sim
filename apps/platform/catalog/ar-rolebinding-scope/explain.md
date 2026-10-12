The **binding** decides where rights apply, not the role:

| Binding | Role it points to | Rights apply in |
|---|---|---|
| RoleBinding | Role (same namespace) | that namespace |
| RoleBinding | **ClusterRole** | **only the RoleBinding's namespace** |
| ClusterRoleBinding | ClusterRole | every namespace, plus cluster-scoped resources |

A RoleBinding that points at a ClusterRole lets you define common rules once (`view`, `edit`,
your own `pod-reader`) and hand them out namespace by namespace, instead of copying a Role into
every namespace.

```sh
kubectl -n ar-scope-dev create rolebinding ci-pod-reader --clusterrole=pod-reader --serviceaccount=ar-scope-dev:ci
kubectl auth can-i list pods -n ar-scope-dev  --as=system:serviceaccount:ar-scope-dev:ci   # yes
kubectl auth can-i list pods -n ar-scope-prod --as=system:serviceaccount:ar-scope-dev:ci   # no
```

The trap is `kubectl create clusterrolebinding`, which grants it in every namespace. Note the
`--serviceaccount=<namespace>:<name>` form, and the `system:serviceaccount:<ns>:<name>` user
name when testing with `--as`.
