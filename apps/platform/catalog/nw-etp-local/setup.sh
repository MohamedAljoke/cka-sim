fresh_ns nw-etp-local
fresh_course
kubectl -n nw-etp-local create deployment web --image=nginx:1.27-alpine --dry-run=client -o yaml |
  sed 's|^    spec:$|    spec:\n      nodeSelector:\n        kubernetes.io/hostname: cka-sim-worker2|' | kubectl apply -f - >/dev/null
kubectl -n nw-etp-local create service nodeport web --tcp=80:80 --node-port=30090 --dry-run=client -o yaml |
  sed 's|^spec:$|spec:\n  externalTrafficPolicy: Local|' | kubectl apply -f - >/dev/null
kubectl -n nw-etp-local rollout status deployment/web --timeout=120s >/dev/null
