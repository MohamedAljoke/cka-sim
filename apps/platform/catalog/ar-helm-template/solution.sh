helm template docs web --repo http://cka-sim-charts --version 1.0.0 -n ar-helm-template > /opt/course/ar-helm-template/docs.yaml
kubectl -n ar-helm-template apply -f /opt/course/ar-helm-template/docs.yaml >/dev/null
