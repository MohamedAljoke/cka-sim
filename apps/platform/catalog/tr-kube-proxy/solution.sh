# kubectl -n kube-system get ds kube-proxy -> DESIRED 0: no node matches its nodeSelector
# kubectl -n kube-system get ds kube-proxy -o yaml | grep -A3 nodeSelector -> kube-proxy: enabled
kubectl -n kube-system patch daemonset kube-proxy --type=json -p '[{"op":"remove","path":"/spec/template/spec/nodeSelector/kube-proxy"}]'
kubectl -n kube-system rollout status daemonset kube-proxy --timeout=120s
