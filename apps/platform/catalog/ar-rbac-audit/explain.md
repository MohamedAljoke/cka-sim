RBAC has no "list a user's rights" object: rights come from every binding that names the user,
or one of their groups, as a subject. Search all of them:

```sh
kubectl get rolebindings,clusterrolebindings -A -o wide | grep contractor
kubectl get rolebindings,clusterrolebindings -A -o json |
  jq -r '.items[] | select(any(.subjects[]?; .kind=="User" and .name=="contractor")) | "\(.kind) \(.metadata.namespace // "-") \(.metadata.name) -> \(.roleRef.name)"'
kubectl auth can-i --list -n ar-audit-app --as=contractor
```

Remove the excess carefully. A binding can list several subjects: deleting `deploy-helpers` or
`oncall` outright would also take rights away from the ServiceAccount and the `oncall` group.
Remove only the `contractor` entry from their `subjects` with `kubectl edit`, or by name with jq.
Don't patch by array index: kubectl doesn't keep the subjects in the order you passed the flags. Delete a binding only when the user is its only subject, like `contractor-debug`.

Rights add up, and there is no deny. One forgotten ClusterRoleBinding to `cluster-admin` makes every
other restriction irrelevant, which is why audits always start there. `-o wide` shows the users,
groups and ServiceAccounts of each binding at a glance.
