fresh_ns wl-psa
kubectl -n wl-psa create deployment web --image=busybox:1.36 --replicas=2 -- sleep 1d >/dev/null
kubectl -n wl-psa rollout status deployment/web --timeout=120s >/dev/null
