fresh_ns ar-resourcenames
kubectl -n ar-resourcenames create serviceaccount toggler >/dev/null
kubectl -n ar-resourcenames create configmap feature-flags --from-literal=dark-mode=off >/dev/null
kubectl -n ar-resourcenames create configmap db-config --from-literal=host=db >/dev/null
