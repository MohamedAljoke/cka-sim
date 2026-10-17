kubectl delete storageclass fast-local --ignore-not-found >/dev/null
kubectl annotate storageclass standard storageclass.kubernetes.io/is-default-class=true --overwrite >/dev/null
