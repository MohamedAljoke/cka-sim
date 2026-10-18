kubectl get rolebindings,clusterrolebindings -A -o json |
  jq -r '.items[] | select(any(.subjects[]?; .kind=="User" and .name=="contractor")) | .metadata.name' > /opt/course/ar-rbac-audit/bindings.txt
kubectl delete clusterrolebinding contractor-debug
for b in ar-audit-app/deploy-helpers ar-audit-ops/oncall; do
  kubectl -n "${b%/*}" get rolebinding "${b#*/}" -o json |
    jq 'del(.subjects[] | select(.kind=="User" and .name=="contractor"))' | kubectl replace -f -
done
kubectl auth can-i --list -n ar-audit-app --as=contractor
