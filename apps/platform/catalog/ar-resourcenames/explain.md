`resourceNames` narrows a rule from "all ConfigMaps" to specific objects:

```yaml
rules:
- apiGroups: [""]
  resources: [configmaps]
  resourceNames: [feature-flags]
  verbs: [get, update]
```

```sh
kubectl -n ar-resourcenames create role flags-editor --verb=get,update --resource=configmaps --resource-name=feature-flags
kubectl auth can-i update configmap/feature-flags -n ar-resourcenames --as=system:serviceaccount:ar-resourcenames:toggler   # yes
kubectl auth can-i update configmap/db-config    -n ar-resourcenames --as=system:serviceaccount:ar-resourcenames:toggler   # no
```

The limits matter: `resourceNames` can't restrict `list`, `watch` or `create`. Those requests carry
no object name, so a `list` rule with `resourceNames` either matches nothing or, in older
versions, failed to restrict at all. Also, `kubectl apply` on an existing object uses `get` and
`patch`. Grant `patch` too if the client applies instead of updating.
