selector() { kubectl -n kube-system get daemonset kube-proxy -o jsonpath='{.spec.template.spec.nodeSelector.kube-proxy}'; }
[ -n "$(selector)" ] || exit 0
kubectl -n kube-system patch daemonset kube-proxy --type=json \
  -p '[{"op":"remove","path":"/spec/template/spec/nodeSelector/kube-proxy"}]' >/dev/null
kubectl -n kube-system rollout status daemonset kube-proxy --timeout=120s >/dev/null
