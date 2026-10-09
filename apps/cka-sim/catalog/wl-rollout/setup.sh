fresh_ns release
fresh_course
kubectl -n release create deployment shop --image=nginx:1.26-alpine --replicas=3 >/dev/null
kubectl -n release rollout status deploy/shop --timeout=180s >/dev/null
kubectl -n release set image deploy/shop nginx=nginx:1.27-alpine >/dev/null
kubectl -n release rollout status deploy/shop --timeout=180s >/dev/null
kubectl -n release set image deploy/shop nginx=nginx:1.27-alpine-broken >/dev/null
