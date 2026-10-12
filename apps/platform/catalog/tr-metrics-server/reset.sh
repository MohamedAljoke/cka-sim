kubectl apply -f /etc/cka-sim/addons/metrics-server.yaml >/dev/null
kubectl -n kube-system rollout status deployment metrics-server --timeout=120s >/dev/null
