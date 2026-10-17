fresh_ns wl-probes
kubectl -n wl-probes create deployment api --image=nginx:1.27-alpine --replicas=2 --port=80 >/dev/null
kubectl -n wl-probes rollout status deployment/api --timeout=120s >/dev/null
