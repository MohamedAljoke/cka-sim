kubectl get storageclass legacy >/dev/null 2>&1 || exit 0
kubectl delete storageclass legacy >/dev/null
kubectl annotate storageclass standard storageclass.kubernetes.io/is-default-class=true --overwrite >/dev/null
