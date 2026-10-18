kubectl get storageclass fast-local >/dev/null 2>&1 || exit 0
kubectl delete storageclass fast-local >/dev/null
kubectl annotate storageclass standard storageclass.kubernetes.io/is-default-class=true --overwrite >/dev/null
