sed -i '/^  replicas:/d' /opt/course/wl-hpa-fight/web.yaml
kubectl apply set-last-applied -f /opt/course/wl-hpa-fight/web.yaml >/dev/null
