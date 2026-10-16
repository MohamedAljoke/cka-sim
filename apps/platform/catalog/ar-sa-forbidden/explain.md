Read the error as a sentence: **who** (`system:serviceaccount:ar-sa:reporter`) can't do **which
verb** (`list`) on **which resource** (`pods` in API group `""`, the core group) **where**
(namespace `ar-sa`). Then follow the chain: ServiceAccount → RoleBinding → Role → rules.

```sh
kubectl -n ar-sa get rolebindings -o wide
kubectl -n ar-sa describe role pod-reader            # verbs: get, watch
kubectl auth can-i --list -n ar-sa --as=system:serviceaccount:ar-sa:reporter
kubectl -n ar-sa edit role pod-reader                # add list
```

`get` reads one named object, `list` reads a whole collection. `kubectl get pods` needs `list`,
and `kubectl get pod x` needs `get`. RBAC rules only add rights, and there is no deny, so the fix
is always to add exactly the missing verb, resource or binding.

The trap is binding `cluster-admin` or `edit` "to make it work". It works, but it fails the least
privilege the task asks for.
