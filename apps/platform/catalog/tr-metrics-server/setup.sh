fresh_course
kubectl delete -f /etc/cka-sim/addons/metrics-server.yaml --ignore-not-found --wait=true >/dev/null 2>&1
grep -v -- '--kubelet-insecure-tls' /etc/cka-sim/addons/metrics-server.yaml > "$COURSE/components.yaml"
chmod 0666 "$COURSE/components.yaml"
