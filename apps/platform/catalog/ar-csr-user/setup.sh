fresh_ns ar-csr-user
fresh_course
kubectl -n ar-csr-user create configmap app-settings --from-literal=mode=prod >/dev/null
kubectl delete csr jane --ignore-not-found >/dev/null
kubectl config delete-context jane >/dev/null 2>&1 || true
kubectl config delete-user jane >/dev/null 2>&1 || true
