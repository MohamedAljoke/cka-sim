sed -i 's|^        - --metric-resolution=15s$|&\n        - --kubelet-insecure-tls|' /opt/course/tr-metrics-server/components.yaml
kubectl apply -f /opt/course/tr-metrics-server/components.yaml >/dev/null
kubectl -n kube-system rollout status deployment metrics-server --timeout=120s >/dev/null
