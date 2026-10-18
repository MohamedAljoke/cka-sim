cm() { kubectl -n kube-system get configmap coredns -o yaml; }
cm | grep -qE '^[[:space:]]+log[[:space:]]*$' || exit 0
cm | sed -E '/^[[:space:]]+log[[:space:]]*$/d' | kubectl replace -f - >/dev/null
kubectl -n kube-system rollout restart deployment coredns >/dev/null
kubectl -n kube-system rollout status deployment coredns --timeout=120s >/dev/null
