kubectl delete clusterrolebinding contractor-debug --ignore-not-found >/dev/null
fresh_ns ar-audit-app
fresh_ns ar-audit-ops
fresh_course
kubectl -n ar-audit-app create rolebinding contractor-view --clusterrole=view --user=contractor >/dev/null
kubectl -n ar-audit-app create rolebinding deploy-helpers --clusterrole=edit --user=contractor --serviceaccount=ar-audit-app:default >/dev/null
kubectl -n ar-audit-ops create rolebinding oncall --clusterrole=admin --group=oncall --user=contractor >/dev/null
kubectl create clusterrolebinding contractor-debug --clusterrole=cluster-admin --user=contractor >/dev/null
