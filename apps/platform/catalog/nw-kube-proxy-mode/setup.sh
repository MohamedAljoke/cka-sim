fresh_ns nw-kube-proxy-mode
fresh_course
kubectl -n nw-kube-proxy-mode create deployment api --image=nginx:1.27-alpine --replicas=2 >/dev/null
kubectl -n nw-kube-proxy-mode expose deployment api --port=80 >/dev/null
kubectl -n nw-kube-proxy-mode rollout status deployment/api --timeout=120s >/dev/null
programmed() { iptables-save -t nat | grep -q 'nw-kube-proxy-mode/api'; }
wait_for 30 programmed
