fresh_ns tr-debug
fresh_course
kubectl -n tr-debug run app --image=registry.k8s.io/pause:3.10 >/dev/null
kubectl -n tr-debug wait --for=condition=Ready pod/app --timeout=120s >/dev/null
