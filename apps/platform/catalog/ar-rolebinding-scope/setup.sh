fresh_ns ar-scope-dev
fresh_ns ar-scope-prod
kubectl -n ar-scope-dev create serviceaccount ci >/dev/null
kubectl delete clusterrole pod-reader --ignore-not-found >/dev/null
kubectl create clusterrole pod-reader --verb=get,list,watch --resource=pods >/dev/null
for b in $(kubectl get clusterrolebindings -o jsonpath='{range .items[?(@.roleRef.name=="pod-reader")]}{.metadata.name}{" "}{end}'); do
  kubectl delete clusterrolebinding "$b" >/dev/null
done
