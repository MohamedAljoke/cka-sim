fresh_ns ar-raw-api
fresh_course
kubectl -n ar-raw-api create deployment web --image=nginx:1.27-alpine --replicas=3 >/dev/null
kubectl -n ar-raw-api rollout status deployment/web --timeout=120s >/dev/null
