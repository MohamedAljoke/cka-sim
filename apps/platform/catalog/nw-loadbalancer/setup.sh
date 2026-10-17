fresh_ns nw-loadbalancer
fresh_course
kubectl -n nw-loadbalancer create deployment shop --image=nginx:1.27-alpine >/dev/null
kubectl -n nw-loadbalancer rollout status deployment/shop --timeout=120s >/dev/null
